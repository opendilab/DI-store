package node_tracker

import (
	"context"
	pbNodeTracker "di_store/pb/node_tracker"
	"di_store/tracing"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

type NodeTrackerClient struct {
	rpcConn   *grpc.ClientConn
	rpcClient pbNodeTracker.NodeTrackerClient
}

func NewNodeTrackerClient(nodeTracker string) (*NodeTrackerClient, error) {
	conn, err := grpc.Dial(nodeTracker, append(tracing.GrpcDialOption, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(dialTimeout))...)
	if err != nil {
		return nil, fmt.Errorf("dial node tracker %v: %v", nodeTracker, err)
	}
	client := pbNodeTracker.NewNodeTrackerClient(conn)
	return &NodeTrackerClient{rpcConn: conn, rpcClient: client}, nil
}

func (client *NodeTrackerClient) RegisterStorageServer(
	hostname string, rpcPort int, plasmaSocket string) (
	*pbNodeTracker.RegisterStorageServerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterStorageServer(ctx, &pbNodeTracker.StorageServer{
		Hostname:     hostname,
		RpcPort:      int32(rpcPort),
		PlasmaSocket: plasmaSocket,
	})
	if err != nil {
		err = fmt.Errorf("RegisterStorageServer: %v", err)
	}
	return r, err
}

func (client *NodeTrackerClient) RegisterStorageClient(serverHostname string) (*pbNodeTracker.RegisterStorageClientResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterStorageClient(ctx, &pbNodeTracker.StorageClient{ServerHostname: serverHostname})
	if err != nil {
		err = fmt.Errorf("RegisterStorageClient: %v", err)
	}
	return r, err
}

func (client *NodeTrackerClient) RegisterObject(ctx context.Context, objID, server string) (*pbNodeTracker.RegisterObjectResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTrackerClient.RegisterObject")
	defer span.Finish()
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	r, err := client.rpcClient.RegisterObject(ctx, &pbNodeTracker.RegisterObjectRequest{ObjectIdHex: objID, ServerHostname: server})
	if err != nil {
		err = fmt.Errorf("RegisterObject: %v", err)
	}
	return r, err
}

func (client *NodeTrackerClient) ServerInfo(ctx context.Context, serverList ...string) ([]*pbNodeTracker.StorageServer, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	r, err := client.rpcClient.ServerInfo(ctx, &pbNodeTracker.ServerInfoRequest{ServerHostnameList: serverList})
	if err != nil {
		err = fmt.Errorf("ServerInfo: %v", err)
	}
	return r.StorageServerList, nil
}

func (client *NodeTrackerClient) ObjectInfo(ctx context.Context, objList ...string) ([]*pbNodeTracker.ObjectInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	r, err := client.rpcClient.ObjectInfo(ctx, &pbNodeTracker.ObjectInfoRequest{ObjectIdHexList: objList})
	if err != nil {
		err = fmt.Errorf("ObjectInfo: %v", err)
	}
	return r.ObjectInfoList, nil
}

func (client *NodeTrackerClient) ObjectDelete(objList ...string) (*pbNodeTracker.ObjectDeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	r, err := client.rpcClient.ObjectDelete(ctx, &pbNodeTracker.ObjectDeleteRequest{ObjectIdHexList: objList})
	if err != nil {
		err = fmt.Errorf("ObjectDelete: %v", err)
	}
	return r, nil
}
