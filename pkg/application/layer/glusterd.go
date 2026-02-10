package layer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/model"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Glusterd struct {
	capability.Logger
	GlusterManager model.GlusterVolumeManager
}

type GlusterdState struct {
	Volumes model.GlusterVolumes
}

func NewGlusterd(logger capability.Logger, manager model.GlusterVolumeManager) *Glusterd {
	return &Glusterd{
		Logger:         logger,
		GlusterManager: manager,
	}
}

func (s *Glusterd) Up(ctx context.Context, state *State) error {
	mountedState := state.Mounted
	if mountedState == nil {
		return ErrInvalidState
	}

	fields := map[string]interface{}{
		"device":      mountedState.Device,
		"mountpoints": len(mountedState.Mountpoints),
	}

	rawEndpoints := strings.TrimSpace(os.Getenv("ETCD_ENDPOINTS"))
	if rawEndpoints == "" {
		err := fmt.Errorf("ETCD_ENDPOINTS must be set (comma-separated)")
		fields["error"] = err
		s.Error("Missing ETCD configuration", fields)
		return err
	}

	endpoints := strings.Split(rawEndpoints, ",")

	fields["etcd_endpoints"] = endpoints

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		fields["error"] = err
		s.Error("Failed to create etcd client", fields)
		return err
	}
	defer func() { _ = etcdClient.Close() }()

	etcdSession, err := concurrency.NewSession(etcdClient)
	if err != nil {
		fields["error"] = err
		s.Error("Failed to create etcd session", fields)
		return err
	}
	defer func() { _ = etcdSession.Close() }()

	var volumes model.GlusterVolumes
	for _, mountpoint := range mountedState.Mountpoints {
		volumeName := mountpoint.Label
		brickPath := mountpoint.Path + "/brick"
		mountPoint := filepath.Join("/mnt/gluster", volumeName)

		fields["volume"] = volumeName
		fields["brick_path"] = brickPath
		fields["mount_point"] = mountPoint

		s.Info("Creating Gluster volume", fields)

		hostname, err := os.Hostname()
		if err != nil {
			fields["error"] = err
			s.Error("Failed to get hostname", fields)
			return err
		}

		if err := os.MkdirAll(brickPath, 0o755); err != nil {
			fields["error"] = err
			s.Error("Failed to create brick directory", fields)
			return fmt.Errorf("failed to create brick directory: %w", err)
		}

		readyCtx, cancelReady := context.WithTimeout(ctx, 10*time.Minute)
		defer cancelReady()

		if err := etcdMarkReady(readyCtx, etcdClient, "/gluster", volumeName, hostname, brickPath); err != nil {
			fields["error"] = err
			s.Error("Failed to mark node ready in etcd", fields)
			return err
		}

		if err := etcdWaitForReady(readyCtx, etcdClient, "/gluster", volumeName, len(endpoints)); err != nil {
			fields["error"] = err
			s.Error("Timeout waiting for all nodes to become ready", fields)
			return err
		}

		mutex := concurrency.NewMutex(etcdSession, "/gluster/mutex/"+volumeName)
		lockCtx, cancelLock := context.WithTimeout(ctx, 10*time.Minute)
		defer cancelLock()

		if err := mutex.Lock(lockCtx); err != nil {
			fields["error"] = err
			s.Error("Failed to acquire distributed lock", fields)
			return err
		}
		s.Info("Acquired distributed lock for volume operations", fields)
		if err := func() error {
			defer func() { _ = mutex.Unlock(context.Background()) }()

			exists := s.GlusterManager.VolumeExists(volumeName)
			if !exists {
				s.Info("Creating Gluster volume (leader via lock)", fields)
				if err := s.GlusterManager.CreateVolume(volumeName, brickPath); err != nil {
					return err
				}
			}

			started := s.GlusterManager.VolumeStarted(volumeName)
			if !started {
				s.Info("Starting Gluster volume (leader via lock)", fields)
				if err := s.GlusterManager.StartVolume(volumeName); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			fields["error"] = err
			s.Error("Failed to create/start volume within distributed lock", fields)
			return err
		}

		waitCtx, cancelWait := context.WithTimeout(ctx, 10*time.Minute)
		defer cancelWait()
		if err := waitForVolumeReady(waitCtx, s.GlusterManager, volumeName); err != nil {
			fields["error"] = err
			s.Error("Volume did not become ready in time", fields)
			return err
		}

		// Mount the volume via FUSE so the host can access it through the shared bind mount.
		err = s.GlusterManager.MountVolume(volumeName, mountPoint)
		if err != nil {
			fields["error"] = err
			s.Error("Failed to mount Gluster volume via FUSE", fields)
			return err
		}

		volumes = append(volumes, model.GlusterVolume{
			Name:      volumeName,
			BrickPath: brickPath,
			Started:   true,
		})

		s.Info("Gluster volume created, started, and mounted successfully", fields)
	}

	state.Glusterd = &GlusterdState{
		Volumes: volumes,
	}

	return nil
}

func etcdMarkReady(ctx context.Context, cli *clientv3.Client, prefix, volumeName, nodeID, brickPath string) error {
	lease, err := cli.Grant(ctx, 600)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/ready/%s/%s", prefix, volumeName, nodeID)
	_, err = cli.Put(ctx, key, brickPath, clientv3.WithLease(lease.ID))
	return err
}

func etcdWaitForReady(ctx context.Context, cli *clientv3.Client, prefix, volumeName string, expectedNodes int) error {
	keyPrefix := fmt.Sprintf("%s/ready/%s/", prefix, volumeName)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		resp, err := cli.Get(ctx, keyPrefix, clientv3.WithPrefix())
		if err != nil {
			return err
		}
		if len(resp.Kvs) >= expectedNodes {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func waitForVolumeReady(ctx context.Context, mgr model.GlusterVolumeManager, volumeName string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		exists := mgr.VolumeExists(volumeName)
		if exists {
			started := mgr.VolumeStarted(volumeName)
			if started {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Glusterd) Down(ctx context.Context, state *State) error {
	return nil
}

func (s *Glusterd) String() string {
	return "glusterd"
}

func (s *Glusterd) Dependencies() []string {
	return []string{"mounted"}
}
