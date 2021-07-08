package node_tracker

import (
	"context"
	pbNodeTracker "di_store/pb/node_tracker"
	pbObjectStore "di_store/pb/storage_server"
	"di_store/tracing"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"go.uber.org/multierr"
)

type NodeTracker struct {
	pbNodeTracker.UnimplementedNodeTrackerServer
	EtcdClient *clientv3.Client

	EtcdHost string
	EtcdPort int
	RpcPort  int
}

func NewNodeTracker(etcdHost string, etcdPort, rpcPort int) (*NodeTracker, error) {
	cfg := clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{fmt.Sprint(etcdHost, ":", etcdPort)},
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &NodeTracker{EtcdClient: client, EtcdHost: etcdHost, EtcdPort: etcdPort, RpcPort: rpcPort}, nil
}

func getIPFromCtx(ctx context.Context) (ip string) {
	p, _ := peer.FromContext(ctx)
	switch addr := p.Addr.(type) {
	case *net.UDPAddr:
		ip = addr.IP.String()
	case *net.TCPAddr:
		ip = addr.IP.String()
	}
	return
}

func lastToken(s string) string {
	idx := strings.LastIndex(s, "/")
	return s[idx+1:]
}

func (tracker *NodeTracker) EtcdGetStorageServer(ctx context.Context, hostname string) (*pbNodeTracker.StorageServer, error) {
	prefix := fmt.Sprintf("/storage_server/%s/", hostname)
	result, err := tracker.EtcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	plasmaSocket := ""
	rpcTarget := ""
	for _, item := range result.Kvs {
		fullKey := string(item.Key)
		switch lastToken(fullKey) {
		case "plasma_socket":
			plasmaSocket = string(item.Value)
		case "rpc_target":
			rpcTarget = string(item.Value)
		}
	}
	return &pbNodeTracker.StorageServer{Hostname: hostname, RpcTarget: rpcTarget, PlasmaSocket: plasmaSocket}, nil
}

func (tracker *NodeTracker) EtcdGetAllStorageServer(ctx context.Context) ([]*pbNodeTracker.StorageServer, error) {
	result, err := tracker.EtcdClient.Get(ctx, "/storage_server/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	prevHostname := ""
	plasmaSocket := ""
	rpcTarget := ""
	var serverList []*pbNodeTracker.StorageServer

	for _, item := range result.Kvs {
		fullKey := strings.Split(string(item.Key), "/")
		if len(fullKey) != 4 {
			log.Fatalf("etcd record error: %v", item.Key)
		}
		hostname, key := fullKey[len(fullKey)-2], fullKey[len(fullKey)-1]
		if hostname != prevHostname {
			if prevHostname != "" {
				serverList = append(serverList,
					&pbNodeTracker.StorageServer{
						Hostname:     prevHostname,
						PlasmaSocket: plasmaSocket,
						RpcTarget:    rpcTarget})
			}
			prevHostname = hostname
		}
		switch key {
		case "plasma_socket":
			plasmaSocket = string(item.Value)
		case "rpc_target":
			rpcTarget = string(item.Value)
		}
	}
	if prevHostname != "" {
		serverList = append(serverList,
			&pbNodeTracker.StorageServer{
				Hostname:     prevHostname,
				PlasmaSocket: plasmaSocket,
				RpcTarget:    rpcTarget})
	}
	return serverList, nil
}

func (tracker *NodeTracker) ServerInfo(ctx context.Context, in *pbNodeTracker.ServerInfoRequest) (*pbNodeTracker.ServerInfoResponse, error) {
	hostnameList := in.GetServerHostnameList()
	var serverList []*pbNodeTracker.StorageServer

	if len(hostnameList) == 0 {
		var err error
		serverList, err = tracker.EtcdGetAllStorageServer(ctx)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	} else {
		for _, hostname := range hostnameList {
			serverInfo, err := tracker.EtcdGetStorageServer(ctx, hostname)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			serverList = append(serverList, serverInfo)
		}
	}
	return &pbNodeTracker.ServerInfoResponse{StorageServerList: serverList}, nil
}
func (tracker *NodeTracker) EtcdGetObjectInfo(ctx context.Context, objIDList []string) ([]*pbNodeTracker.ObjectInfo, error) {
	objInfoList := make([]*pbNodeTracker.ObjectInfo, len(objIDList))
	for i, objId := range objIDList {
		result, err := tracker.EtcdClient.Get(ctx, fmt.Sprintf("/object/%s/", objId), clientv3.WithPrefix())
		if err != nil {
			return nil, err
		}
		hostnameList := make([]string, len(result.Kvs))
		for j, item := range result.Kvs {
			serverHostname := lastToken(string(item.Key))
			hostnameList[j] = serverHostname

		}
		objInfoList[i] = &pbNodeTracker.ObjectInfo{ObjectIdHex: objId, ServerHostnameList: hostnameList}
	}
	return objInfoList, nil
}

func (tracker *NodeTracker) EtcdObjectDelete(ctx context.Context, objIDList []string) error {
	ops := make([]clientv3.Op, 0, len(objIDList))
	for _, objID := range objIDList {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/object/%s/", objID), clientv3.WithPrefix()))
	}
	_, err := tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
	return err
}

func (tracker *NodeTracker) EtcdGetRpcTarget(ctx context.Context, server string) (string, error) {
	result, err := tracker.EtcdClient.Get(ctx, fmt.Sprintf("/storage_server/%s/rpc_target", server))
	if err != nil {
		return "", err
	}
	if len(result.Kvs) == 0 {
		err := fmt.Errorf("RPC Target of %s not found", server)
		log.Error(err)
		return "", err
	}
	return string(result.Kvs[0].Value), nil
}

func (tracker *NodeTracker) ObjectInfo(ctx context.Context, in *pbNodeTracker.ObjectInfoRequest) (*pbNodeTracker.ObjectInfoResponse, error) {
	start := time.Now()
	objIDList := in.GetObjectIdHexList()
	objInfoList, err := tracker.EtcdGetObjectInfo(ctx, objIDList)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Trace("ObjectInfo duration:", time.Since(start))
	return &pbNodeTracker.ObjectInfoResponse{ObjectInfoList: objInfoList}, nil
}

func (tracker *NodeTracker) objectDeleteOnServer(ctx context.Context, server string, objIDList []string) error {
	log.Debug("object_delete_on_server: ", server, " object list: ", objIDList)
	// todo cache rpcTarget
	rpcTarget, err := tracker.EtcdGetRpcTarget(ctx, server)
	if err != nil {
		return err
	}
	conn, err := grpc.Dial(rpcTarget, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	defer conn.Close()
	c := pbObjectStore.NewObjectStoreClient(conn)
	_, err = c.Delete(ctx, &pbObjectStore.DeleteRequest{ObjectIdHexList: objIDList})
	return err
}

func (tracker *NodeTracker) ObjectDelete(ctx context.Context, in *pbNodeTracker.ObjectDeleteRequest) (*pbNodeTracker.ObjectDeleteResponse, error) {
	objIDList := in.GetObjectIdHexList()
	objInfoList, err := tracker.EtcdGetObjectInfo(ctx, objIDList)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	serverObjMap := make(map[string][]string)
	for _, item := range objInfoList {
		for _, server := range item.ServerHostnameList {
			serverObjMap[server] = append(serverObjMap[server], item.ObjectIdHex)
		}
	}

	log.Debug("serverObjMap:", serverObjMap)

	resultQ := make(chan error, len(serverObjMap))
	for server, objList := range serverObjMap {
		go func(serverHostname string, objIDList []string) {
			resultQ <- tracker.objectDeleteOnServer(ctx, serverHostname, objIDList)
		}(server, objList)
	}

	for i := 0; i < len(serverObjMap); i++ {
		err = multierr.Append(err, <-resultQ)
	}
	// delete object info in etcd anyway
	err = multierr.Append(err, tracker.EtcdObjectDelete(ctx, objIDList))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &pbNodeTracker.ObjectDeleteResponse{}, nil
}

func (tracker *NodeTracker) RegisterObject(ctx context.Context, in *pbNodeTracker.RegisterObjectRequest) (*pbNodeTracker.RegisterObjectResponse, error) {
	objID := in.GetObjectIdHex()
	serverHostname := in.GetServerHostname()
	key := fmt.Sprint("/object/", objID, "/", serverHostname)

	{ // etcd ctx
		span, ctx := opentracing.StartSpanFromContext(ctx, "etcd.client.put")
		_, err := tracker.EtcdClient.Put(ctx, key, "0")
		span.Finish()
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	return &pbNodeTracker.RegisterObjectResponse{}, nil
}

func (tracker *NodeTracker) RegisterStorageClient(ctx context.Context, in *pbNodeTracker.StorageClient) (*pbNodeTracker.RegisterStorageClientResponse, error) {
	server := in.GetServerHostname()
	serverInfo, err := tracker.EtcdGetStorageServer(ctx, server)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &pbNodeTracker.RegisterStorageClientResponse{
		EtcdServer:    &pbNodeTracker.EtcdServer{IpAddr: tracker.EtcdHost, Port: int32(tracker.EtcdPort)},
		StorageServer: serverInfo,
	}, nil
}

func (tracker *NodeTracker) RegisterStorageServer(
	ctx context.Context, in *pbNodeTracker.StorageServer) (*pbNodeTracker.RegisterStorageServerResponse, error) {
	storageServerIP := getIPFromCtx(ctx)
	if in.GetIpAddr() != "" {
		storageServerIP = in.GetIpAddr()
	}

	hostname := in.GetHostname()
	rpcPort := in.GetRpcPort()
	plasmaSocket := in.GetPlasmaSocket()
	keyPrefix := fmt.Sprint("/storage_server/", hostname)
	items := map[string]string{
		"state":         "up",
		"rpc_target":    fmt.Sprint(storageServerIP, ":", rpcPort),
		"plasma_socket": plasmaSocket,
	}
	{ // etcd ctx
		span, ctx := opentracing.StartSpanFromContext(ctx, "etcd.put")
		ops := make([]clientv3.Op, 0, len(items))
		for k, v := range items {
			ops = append(ops, clientv3.OpPut(fmt.Sprint(keyPrefix, "/", k), v))
		}
		_, err := tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
		span.Finish()

		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	return &pbNodeTracker.RegisterStorageServerResponse{
		EtcdServer:          &pbNodeTracker.EtcdServer{IpAddr: tracker.EtcdHost, Port: int32(tracker.EtcdPort)},
		StorageServerIpAddr: storageServerIP,
	}, nil
}

func (tracker *NodeTracker) Serve() {
	listenAddr := fmt.Sprint(":", tracker.RpcPort)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Info("start node tracker, listening on ", listenAddr)

	grpcServer := grpc.NewServer(tracing.GrpcServerOption...)
	pbNodeTracker.RegisterNodeTrackerServer(grpcServer, tracker)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
