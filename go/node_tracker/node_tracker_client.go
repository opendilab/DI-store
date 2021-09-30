package node_tracker

import (
	"context"
	pbNodeTracker "di_store/pb/node_tracker"
	"di_store/util"

	"github.com/opentracing/opentracing-go"
	ot_log "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type NodeTrackerClient struct {
	rpcConn   *grpc.ClientConn
	rpcClient pbNodeTracker.NodeTrackerClient
}

func NewNodeTrackerClient(ctx context.Context, nodeTracker string) (*NodeTrackerClient, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NewNodeTrackerClient.Dial")
	defer span.Finish()
	conn, err := grpc.DialContext(ctx, nodeTracker, util.GrpcDialOption()...)
	if err != nil {
		span.LogFields(ot_log.Error(err))
		return nil, errors.Wrapf(err, "dial node tracker: %v", nodeTracker)
	}
	client := pbNodeTracker.NewNodeTrackerClient(conn)
	return &NodeTrackerClient{rpcConn: conn, rpcClient: client}, nil
}

func (client *NodeTrackerClient) RegisterStorageServer(
	hostname string, rpcPort int, plasmaSocket string, objTransferPort int, groupList []string) (
	*pbNodeTracker.RegisterStorageServerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterStorageServer(ctx, &pbNodeTracker.StorageServer{
		Hostname:        hostname,
		RpcPort:         int32(rpcPort),
		PlasmaSocket:    plasmaSocket,
		ObjTransferPort: int32(objTransferPort),
		GroupList:       groupList,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.RegisterStorageServer")
	}
	return r, nil
}

func (client *NodeTrackerClient) UnregisterStorageServer(hostname string) (
	*pbNodeTracker.UnregisterStorageServerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.UnregisterStorageServer(ctx, &pbNodeTracker.StorageServer{
		Hostname: hostname,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.UnregisterStorageServer")
	}
	return r, nil
}

func (client *NodeTrackerClient) RegisterStorageClient(serverHostname string) (*pbNodeTracker.RegisterStorageClientResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterStorageClient(ctx, &pbNodeTracker.StorageClient{ServerHostname: serverHostname})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.RegisterStorageClient")
	}
	return r, nil
}

func (client *NodeTrackerClient) RegisterObject(ctx context.Context, objID, server string) (*pbNodeTracker.RegisterObjectResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTrackerClient.RegisterObject")
	defer span.Finish()

	ctx, cancel := context.WithTimeout(ctx, util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterObject(ctx, &pbNodeTracker.RegisterObjectRequest{ObjectIdHex: objID, ServerHostname: server})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.RegisterObject")
	}
	return r, nil
}

func (client *NodeTrackerClient) ServerInfo(ctx context.Context, serverList ...string) ([]*pbNodeTracker.StorageServer, error) {
	ctx, cancel := context.WithTimeout(ctx, util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.ServerInfo(ctx, &pbNodeTracker.ServerInfoRequest{ServerHostnameList: serverList})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.ServerInfo")
	}
	return r.StorageServerList, nil
}

func (client *NodeTrackerClient) ObjectInfo(ctx context.Context, objList ...string) ([]*pbNodeTracker.ObjectInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTrackerClient.ObjectInfo")
	defer span.Finish()

	r, err := client.rpcClient.ObjectInfo(ctx, &pbNodeTracker.ObjectInfoRequest{ObjectIdHexList: objList})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.ObjectInfo")
	}
	return r.ObjectInfoList, nil
}

func (client *NodeTrackerClient) ObjectDelete(objList ...string) (*pbNodeTracker.ObjectDeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), util.CommonConfig.RequestTimeout)
	defer cancel()
	r, err := client.rpcClient.ObjectDelete(ctx, &pbNodeTracker.ObjectDeleteRequest{ObjectIdHexList: objList})
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTrackerClient.ObjectDelete")
	}
	return r, nil
}
