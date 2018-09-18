package fc

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/mathu97/csi-connectors/fibrechannel"
	"golang.org/x/net/context"
	"fmt"
)

type fcNodeServer struct {
	*csicommon.DefaultNodeServer
	Driver *CSIDriver
}

var fcDisk fcDevice

func (ns *fcNodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *fcNodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (ns *fcNodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	fcDevice, err := getFCInfo(req)

	if err != nil {
		return nil, err
	}
	fcDisk = *fcDevice
	disk, connectError := fibrechannel.Connect(*fcDevice.connector, nil)

	fcDisk.disk = disk

	//Mount
	fcmounter := getFCDiskMounter(req)
	MountDisk(*fcmounter, disk)

	if connectError != nil {
		return nil, connectError
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *fcNodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	if err := fibrechannel.Disconnect(fcDisk.disk, nil); err != nil {
		return nil, err
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *fcNodeServer) NodeGetCapabilities(context.Context, *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_UNKNOWN,
					},
				},
			},
		},
	}, nil
}

func (ns *fcNodeServer) NodeGetId(ctx context.Context, req *csi.NodeGetIdRequest) (*csi.NodeGetIdResponse, error) {
	return &csi.NodeGetIdResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

func (ns *fcNodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ns.Driver.nodeID,
	}, nil
}

//MountDisk mounts the fibre channel disk with the given parameters
func MountDisk(mnter FCMounter, devicePath string) error {
	mntPath := mnter.TargetPath
	notMnt, err := mnter.Mounter.IsLikelyNotMountPoint(mntPath)

	if err != nil {
		return fmt.Errorf("Heuristic determination of mount point failed: %v", err)
	}

	if !notMnt {
		fmt.Printf("fc: %s already mounted", mnter.TargetPath)
	}

	if err = mnter.Mounter.FormatAndMount(devicePath, mnter.TargetPath, mnter.FsType, nil); err != nil {
		return fmt.Errorf("fc: failed to mount fc volume %s [%s] to %s, error %v", devicePath, mnter.FsType, mnter.TargetPath, err)
	}

	return nil
}