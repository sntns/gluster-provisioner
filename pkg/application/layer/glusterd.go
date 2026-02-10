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
	if state.Mounted == nil {
		return ErrInvalidState
	}

	cli, sess, err := initEtcd()
	if err != nil {
		return err
	}
	defer cli.Close()
	defer sess.Close()

	var vols model.GlusterVolumes

	for _, mp := range state.Mounted.Mountpoints {
		v, err := s.reconcileVolume(
			ctx,
			cli,
			sess,
			mp,
			len(state.Mounted.Mountpoints),
		)
		if err != nil {
			return err
		}

		vols = append(vols, v)
	}

	state.Glusterd = &GlusterdState{
		Volumes: vols,
	}

	return nil
}

func initEtcd() (*clientv3.Client, *concurrency.Session, error) {
	raw := strings.TrimSpace(os.Getenv("ETCD_ENDPOINTS"))
	if raw == "" {
		return nil, nil, fmt.Errorf("ETCD_ENDPOINTS must be set")
	}

	endpoints := strings.Split(raw, ",")

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, nil, err
	}

	sess, err := concurrency.NewSession(cli)
	if err != nil {
		cli.Close()
		return nil, nil, err
	}

	return cli, sess, nil
}

func (s *Glusterd) reconcileVolume(ctx context.Context, cli *clientv3.Client, sess *concurrency.Session, mp model.Mountpoint, expectedNodes int) (model.GlusterVolume, error) {

	volume := mp.Label
	brick := mp.Path + "/brick"
	mount := filepath.Join("/mnt/gluster", volume)

	fields := map[string]any{
		"volume": volume,
	}

	host, _ := os.Hostname()

	if err := os.MkdirAll(brick, 0755); err != nil {
		return model.GlusterVolume{}, err
	}

	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := etcdMarkReady(readyCtx, cli, "/gluster", volume, host, brick); err != nil {
		return model.GlusterVolume{}, err
	}

	if err := etcdWaitForReady(readyCtx, cli, "/gluster", volume, expectedNodes); err != nil {
		return model.GlusterVolume{}, err
	}

	mutex := concurrency.NewMutex(sess, "/gluster/mutex/"+volume)

	lockCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := mutex.Lock(lockCtx); err != nil {
		return model.GlusterVolume{}, err
	}

	func() {
		defer mutex.Unlock(context.Background())

		if !s.GlusterManager.VolumeExists(volume) {
			s.GlusterManager.CreateVolume(volume, brick)
		}

		if !s.GlusterManager.VolumeStarted(volume) {
			s.GlusterManager.StartVolume(volume)
		}
	}()

	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := waitForVolumeReady(waitCtx, s.GlusterManager, volume); err != nil {
		return model.GlusterVolume{}, err
	}

	if err := s.GlusterManager.EnsureMounted(volume, mount); err != nil {
		return model.GlusterVolume{}, err
	}

	s.Info("Volume reconciled", fields)

	return model.GlusterVolume{
		Name:      volume,
		BrickPath: brick,
		Started:   true,
	}, nil
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
