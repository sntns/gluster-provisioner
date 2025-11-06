package disk

import (
	"encoding/json"
	"os/exec"
	"time"
)

type LsblkOutput struct {
	Blockdevices BlockDevices `json:"blockdevices"`
}

type BlockDevices []BlockDevice

type BlockDevice struct {
	Name        string        `json:"name"`
	Kname       string        `json:"kname,omitempty"`
	MajMin      string        `json:"maj:min,omitempty"`
	Fstype      string        `json:"fstype,omitempty"`
	Mountpoints []string      `json:"mountpoints,omitempty"`
	Label       string        `json:"label,omitempty"`
	UUID        string        `json:"uuid,omitempty"`
	Parttype    string        `json:"parttype,omitempty"`
	Partlabel   string        `json:"partlabel,omitempty"`
	Partuuid    string        `json:"partuuid,omitempty"`
	Partflags   string        `json:"partflags,omitempty"`
	Ra          uint64        `json:"ra,omitempty"`
	Ro          bool          `json:"ro,omitempty"`
	Rm          bool          `json:"rm,omitempty"`
	Hotplug     bool          `json:"hotplug,omitempty"`
	Size        uint64        `json:"size,omitempty"`
	State       string        `json:"state,omitempty"`
	Owner       string        `json:"owner,omitempty"`
	Group       string        `json:"group,omitempty"`
	Mode        string        `json:"mode,omitempty"`
	Alignment   uint64        `json:"alignment,omitempty"`
	MinIO       uint64        `json:"min-io,omitempty"`
	OptIO       uint64        `json:"opt-io,omitempty"`
	PhySec      uint64        `json:"phy-sec,omitempty"`
	LogSec      uint64        `json:"log-sec,omitempty"`
	Rota        bool          `json:"rota,omitempty"`
	Sched       string        `json:"sched,omitempty"`
	RQSize      uint64        `json:"rq-size,omitempty"`
	Type        string        `json:"type,omitempty"`
	DiscAln     uint64        `json:"disc-aln,omitempty"`
	DiscGran    uint64        `json:"disc-gran,omitempty"`
	DiscMax     uint64        `json:"disc-max,omitempty"`
	DiscZero    bool          `json:"disc-zero,omitempty"`
	WSame       uint64        `json:"wsame,omitempty"`
	WWN         string        `json:"wwn,omitempty"`
	Serial      string        `json:"serial,omitempty"`
	Children    []BlockDevice `json:"children,omitempty"`
}

func ListBlocks(device string) (*BlockDevice, error) {
	time.Sleep(1 * time.Second)

	cmd := exec.Command("lsblk", "--json", "--bytes", device)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result LsblkOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}
	if len(result.Blockdevices) == 0 {
		return nil, nil
	}
	return &result.Blockdevices[0], nil
}

func ListFilesystems(device string) (*BlockDevice, error) {
	time.Sleep(1 * time.Second)

	cmd := exec.Command("lsblk", "--json", "--bytes", "--fs", device)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var result LsblkOutput
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}
	if len(result.Blockdevices) == 0 {
		return nil, nil
	}
	return &result.Blockdevices[0], nil
}
