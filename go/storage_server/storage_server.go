package storage_server

import (
	"C"
	"context"
	"di_store/node_tracker"
	pbNodeTracker "di_store/pb/node_tracker"
	pbObjectStore "di_store/pb/storage_server"
	"di_store/plasma_client"
	plasma "di_store/plasma_server"
	"di_store/util"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type StorageServer struct {
	pbObjectStore.UnimplementedObjectStoreServer
	Hostname             string
	PlasmaClient         *plasma_client.PlasmaClient
	NodeTrackerRpcTarget string
	// NodeTrackerClient   *node_tracker.NodeTrackerClient
	RpcListener         net.Listener
	ObjTransferListener net.Listener
	ServerInfoMap       map[string]*pbNodeTracker.StorageServer
	mu                  sync.Mutex
	FetchTaskManager    *FetchTaskManager
	GroupList           []string
}

func NewStorageServer(
	ctx context.Context,
	hostname string,
	rpcPort int,
	nodeTrackerIp string,
	nodeTrackerPort int,
	plasmaSocket string,
	plasmaMemoryByte int,
	groupList []string,
) (*StorageServer, error) {
	if hostname == "" {
		hostname = os.Getenv("DI_STORE_NODE_NAME")
	}
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	}
	plasmaSocket = fmt.Sprintf("%s-%s", plasmaSocket, util.RandomHexString(10))
	err := plasma.RunPlasmaStoreServer(ctx, plasmaMemoryByte, plasmaSocket)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run plasma store server")
	}
	log.Info("running PlasmaStoreServer, path: ", plasmaSocket, " ,memory: ", plasmaMemoryByte)

	rpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	rpcPort = rpcLis.Addr().(*net.TCPAddr).Port

	plasmaClient, err := plasma_client.NewClient(plasmaSocket)
	if err != nil {
		return nil, err
	}
	nodeTrackerRpcTarget := fmt.Sprintf("%s:%d", nodeTrackerIp, nodeTrackerPort)
	nodeTrackerClient, err := node_tracker.NewNodeTrackerClient(context.Background(), nodeTrackerRpcTarget)
	if err != nil {
		return nil, err
	}

	objTransferLis, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	objTransferPort := objTransferLis.Addr().(*net.TCPAddr).Port

	_, err = nodeTrackerClient.RegisterStorageServer(hostname, rpcPort, plasmaSocket, objTransferPort, groupList)
	if err != nil {
		return nil, fmt.Errorf("failed to register storage server: %v", err)
	}

	server := &StorageServer{
		Hostname:             hostname,
		PlasmaClient:         plasmaClient,
		NodeTrackerRpcTarget: nodeTrackerRpcTarget,
		RpcListener:          rpcLis,
		ObjTransferListener:  objTransferLis,
		ServerInfoMap:        make(map[string]*pbNodeTracker.StorageServer),
		FetchTaskManager:     NewFetchTaskManager(),
		GroupList:            groupList,
	}

	err = server.UpdateServerInfoMap(context.Background())
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (server *StorageServer) UpdateServerInfoMap(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.UpdateServerRpcTargetMap")
	defer span.Finish()
	server.mu.Lock()
	defer server.mu.Unlock()

	client, err := node_tracker.NewNodeTrackerClient(ctx, server.NodeTrackerRpcTarget)
	if err != nil {
		return errors.Wrapf(err, "StorageServer.UpdateServerInfoMap")
	}
	serverInfo, err := client.ServerInfo(ctx)
	if err != nil {
		return errors.Wrapf(err, "StorageServer.UpdateServerInfoMap")
	}
	server.ServerInfoMap = map[string]*pbNodeTracker.StorageServer{}
	for _, info := range serverInfo {
		server.ServerInfoMap[info.Hostname] = info
	}
	return nil
}

func (server *StorageServer) GetServerInfo(ctx context.Context, hostname string) (*pbNodeTracker.StorageServer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.GetServerInfo")
	defer span.Finish()

	info, exist := server.ServerInfoMap[hostname]
	if !exist {
		err := server.UpdateServerInfoMap(ctx)
		if err != nil {
			return nil, err
		}
		info, exist = server.ServerInfoMap[hostname]
	}
	if !exist {
		return nil, errors.Errorf("StorageServer.GetServerInfo can not find server info of %s", hostname)
	}
	return info, nil
}

func (server *StorageServer) UpdateStorageServer(ctx context.Context, in *pbObjectStore.UpdateStorageServerRequest) (*pbObjectStore.UpdateStorageServerResponse, error) {
	server.mu.Lock()
	defer server.mu.Unlock()

	serverInfo := in.GetServerInfo()
	if serverInfo == nil {
		return nil, errors.Errorf("StorageServer.GetServerInfo can not find server info in request")
	}
	remove := in.GetRemove()
	hostname := serverInfo.GetHostname()
	if remove {
		delete(server.ServerInfoMap, hostname)
	} else {
		server.ServerInfoMap[hostname] = serverInfo
	}

	return &pbObjectStore.UpdateStorageServerResponse{}, nil
}

func (server *StorageServer) Get(ctx context.Context, in *pbObjectStore.GetRequest) (*pbObjectStore.GetResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.Get")
	defer span.Finish()

	oidHex := in.GetObjectIdHex()
	oid, err := hex.DecodeString(oidHex)
	if err != nil {
		return nil, errors.Wrapf(err, "decode object id: %v", string(oidHex))
	}
	data, err := server.PlasmaClient.Get(ctx, oid)
	if err != nil {
		return nil, err
	}

	if data == nil {
		// span.LogFields(ot_log.Int("object_size", -1))
		return &pbObjectStore.GetResponse{NotFound: true}, nil
	} else {
		// span.LogFields(ot_log.Int("object_size", len(data)))
		return &pbObjectStore.GetResponse{Data: data}, nil
	}
}

func oidHexList2oidList(oidHexList []string) ([][]byte, error) {
	oidList := make([][]byte, len(oidHexList))
	for i, oidHex := range oidHexList {
		oid, err := hex.DecodeString(oidHex)
		if err != nil {
			return nil, err
		}
		oidList[i] = oid
	}
	return oidList, nil
}

func (server *StorageServer) Delete(ctx context.Context, in *pbObjectStore.DeleteRequest) (*pbObjectStore.DeleteResponse, error) {
	oidHexList := in.GetObjectIdHexList()
	oidList, err := oidHexList2oidList(oidHexList)
	if err != nil {
		return nil, err
	}
	err = server.PlasmaClient.Delete(ctx, oidList...)
	if err != nil {
		return nil, err
	}
	return &pbObjectStore.DeleteResponse{}, nil
}

func (server *StorageServer) Fetch(ctx context.Context, in *pbObjectStore.FetchRequest) (*pbObjectStore.FetchResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.Fetch")
	defer span.Finish()

	oidHex := in.GetObjectIdHex()
	viaRpc := in.GetViaRpc()
	srcNode := in.GetSrcNode()
	srcNodeOnly := in.GetSrcNodeOnly()

	if srcNodeOnly && srcNode == "" {
		return nil, errors.New("argument error: src_node_only == true && src_node == ''")
	}

	oid, err := hex.DecodeString(oidHex)
	if err != nil {
		return nil, err
	}

	exist, err := server.PlasmaClient.Contains(ctx, oid)
	if err != nil {
		return nil, err
	}

	if exist {
		log.Debugf("object already exists: %v", oidHex)
		return &pbObjectStore.FetchResponse{}, nil
	}

	if viaRpc {
		err = server.FetchTaskManager.Fetch(ctx, oid, srcNode, srcNodeOnly, server.fetchViaRpc)
	} else {
		err = server.FetchTaskManager.Fetch(ctx, oid, srcNode, srcNodeOnly, server.fetchViaSocket)
	}

	if err != nil {
		return nil, err
	}
	return &pbObjectStore.FetchResponse{}, nil
}

func (server *StorageServer) Serve() {
	log.Info("start storage server, listening on port ", (server.RpcListener).Addr())
	go server.objTransferListenLoop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		client, err := node_tracker.NewNodeTrackerClient(context.Background(), server.NodeTrackerRpcTarget)
		if err != nil {
			log.Error("failed to unregister storage server: %v", err)
		}
		_, err = client.UnregisterStorageServer(server.Hostname)
		if err != nil {
			log.Error("failed to unregister storage server: %v", err)
		}
		time.Sleep(time.Second)
		os.Exit(0)
	}()

	grpcServer := grpc.NewServer(util.GrpcServerOption()...)
	pbObjectStore.RegisterObjectStoreServer(grpcServer, server)
	if err := grpcServer.Serve(server.RpcListener); err != nil {
		log.Fatalf("failed to create StorageServer: %+v", err)
	}
}
