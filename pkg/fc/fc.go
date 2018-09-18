package fc

import (
	"encoding/json"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/mathu97/csi-connectors/fibrechannel"
	"k8s.io/kubernetes/pkg/util/mount"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"
)

type fcDevice struct {
	connector *fibrechannel.Connector
	disk      string
}

//FCMounter struct holds required parameters to mount a Fibre Channel Disk
type FCMounter struct {
	ReadOnly     bool
	FsType       string
	MountOptions []string
	Mounter      *mount.SafeFormatAndMount
	Exec         mount.Exec
	DeviceUtil   volumeutil.DeviceUtil
	TargetPath   string
}

func getFCInfo(req *csi.NodePublishVolumeRequest) (*fcDevice, error) {
	volName := req.GetVolumeId()
	lun := req.GetVolumeAttributes()["lun"]
	targetWWNs := req.GetVolumeAttributes()["targetWWNs"]
	wwids := req.GetVolumeAttributes()["WWIDs"]

	if lun == "" || (targetWWNs == "" && wwids == "") {
		return nil, fmt.Errorf("FC target information is missing")
	}

	targetList := []string{}
	wwidList := []string{}

	if err := json.Unmarshal([]byte(targetWWNs), &targetList); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(wwids), &wwidList); err != nil {
		return nil, err
	}

	fcConnector := &fibrechannel.Connector{
		VolumeName: volName,
		TargetWWNs: targetList,
		WWIDs:      wwidList,
		Lun:        lun,
	}

	//Only pass the connector
	return &fcDevice{
		connector: fcConnector,
	}, nil

}

func getFCDiskMounter(req *csi.NodePublishVolumeRequest) *FCMounter {
	readOnly := req.GetReadonly()
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	mountOptions := req.GetVolumeCapability().GetMount().GetMountFlags()
	return &FCMounter{
		ReadOnly:     readOnly,
		FsType:       fsType,
		MountOptions: mountOptions,
		Mounter:      &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
		Exec:         mount.NewOsExec(),
		DeviceUtil:   volumeutil.NewDeviceHandler(volumeutil.NewIOHandler()),
		TargetPath:   req.GetTargetPath(),
	}
}
