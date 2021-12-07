package node_tracker

import (
	"context"
	"di_store/metadata"
	pbNodeTracker "di_store/pb/node_tracker"
	pbStorageServer "di_store/pb/storage_server"
	"di_store/util"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/opentracing/opentracing-go"
	ot_log "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	mvccpb "go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var EtcdTxnLimit = 100

type NodeTracker struct {
	pbNodeTracker.UnimplementedNodeTrackerServer
	EtcdClient        clientv3.KV
	EtcdHost          string
	EtcdPort          int
	RpcHost           string
	RpcPort           int
	fetchTaskQueue    chan func() error
	deleteTaskQueue   chan func() error
	pushInfoTaskQueue chan func() error
	m                 sync.Mutex
}

func NewNodeTracker(etcdHost string, etcdPort int, rpcHost string, rpcPort int) (*NodeTracker, error) {
	cfg := clientv3.Config{
		DialTimeout: util.CommonConfig.DialTimeout,
		Endpoints:   []string{fmt.Sprint(etcdHost, ":", etcdPort)},
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	fetchTaskQueue := make(chan func() error, util.CommonConfig.TaskQueueCap)
	deleteTaskQueue := make(chan func() error, util.CommonConfig.TaskQueueCap)
	pushInfoTaskQueue := make(chan func() error, util.CommonConfig.TaskQueueCap)
	go taskMainLoop(fetchTaskQueue)
	go taskMainLoop(deleteTaskQueue)
	go taskMainLoop(pushInfoTaskQueue)
	return &NodeTracker{
		EtcdClient:        metadata.NewClient(client),
		EtcdHost:          etcdHost,
		EtcdPort:          etcdPort,
		RpcHost:           rpcHost,
		RpcPort:           rpcPort,
		fetchTaskQueue:    fetchTaskQueue,
		deleteTaskQueue:   deleteTaskQueue,
		pushInfoTaskQueue: pushInfoTaskQueue,
	}, nil
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetStorageServer")
	defer span.Finish()

	prefix := fmt.Sprintf("/storage_server/%s/", hostname)
	result, err := tracker.EtcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTracker.EtcdGetStorageServer")
	}
	serverList := buildServerInfoRecordList(result.Kvs)
	if len(serverList) == 0 {
		return nil, errors.Errorf("NodeTracker.EtcdGetStorageServer can not find storage server: %v", hostname)
	}
	return serverList[0], nil
}

func buildServerInfoRecordList(kvs []*mvccpb.KeyValue) []*pbNodeTracker.StorageServer {
	m := map[string]*pbNodeTracker.StorageServer{}
	for _, item := range kvs {
		fullKey := strings.Split(string(item.Key), "/")
		if len(fullKey) != 4 {
			log.Fatalf("EtcdGetStorageServer record error: %v", item.Key)
		}
		hostname, key := fullKey[len(fullKey)-2], fullKey[len(fullKey)-1]
		server, exist := m[hostname]
		if !exist {
			server = &pbNodeTracker.StorageServer{Hostname: hostname}
			m[hostname] = server
		}
		switch key {
		case "ip_addr":
			server.IpAddr = string(item.Value)
		case "plasma_socket":
			server.PlasmaSocket = string(item.Value)
		case "rpc_port":
			i64, err := strconv.ParseInt(string(item.Value), 10, 32)
			if err != nil {
				log.Fatalf("%+v", errors.Wrapf(err, "EtcdGetStorageServer parse error"))
			}
			server.RpcPort = int32(i64)
		case "rpc_target":
			server.RpcTarget = string(item.Value)
		case "obj_transfer_port":
			i64, err := strconv.ParseInt(string(item.Value), 10, 32)
			if err != nil {
				log.Fatalf("%+v", errors.Wrapf(err, "EtcdGetStorageServer parse error"))
			}
			server.ObjTransferPort = int32(i64)
		case "obj_transfer_target":
			server.ObjTransferTarget = string(item.Value)
		case "group_list":
			groupList := string(item.Value)
			if groupList != "" {
				server.GroupList = strings.Split(string(item.Value), ",")
			}
		case "state":
			continue
		default:
			log.Fatalf("%+v", errors.Errorf("EtcdGetStorageServer unknown key: %v", key))
		}

	}

	result := make([]*pbNodeTracker.StorageServer, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func (tracker *NodeTracker) EtcdGetAllStorageServer(ctx context.Context) ([]*pbNodeTracker.StorageServer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetAllStorageServer")
	defer span.Finish()

	result, err := tracker.EtcdClient.Get(ctx, "/storage_server/", clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTracker.EtcdGetAllStorageServer")
	}
	return buildServerInfoRecordList(result.Kvs), nil
}

func (tracker *NodeTracker) ServerInfo(ctx context.Context, in *pbNodeTracker.ServerInfoRequest) (*pbNodeTracker.ServerInfoResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.ServerInfo")
	defer span.Finish()

	hostnameList := in.GetServerHostnameList()
	var serverList []*pbNodeTracker.StorageServer

	if len(hostnameList) == 0 {
		var err error
		serverList, err = tracker.EtcdGetAllStorageServer(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		for _, hostname := range hostnameList {
			serverInfo, err := tracker.EtcdGetStorageServer(ctx, hostname)
			if err != nil {
				return nil, err
			}
			serverList = append(serverList, serverInfo)
		}
	}
	return &pbNodeTracker.ServerInfoResponse{StorageServerList: serverList}, nil
}

func (tracker *NodeTracker) EtcdGetHostnameListOfGroup(ctx context.Context, group string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetHostnameListOfGroup")
	defer span.Finish()
	prefix := fmt.Sprintf("/storage_server_group/%s/", group)
	result, err := tracker.EtcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "NodeTracker.EtcdGetHostnameListOfGroup")
	}

	var groupList []string
	for _, item := range result.Kvs {
		key := string(item.Key)
		group := lastToken(key)
		groupList = append(groupList, group)
	}
	return groupList, nil
}

func (tracker *NodeTracker) EtcdGetObjectList(ctx context.Context, hostname string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetObjectList")
	defer span.Finish()

	result, err := tracker.EtcdClient.Get(ctx, fmt.Sprintf("/hostname2object/%s/", hostname), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTracker.EtcdGetObjectList hostname: %v", hostname)
	}
	objIDList := make([]string, len(result.Kvs))
	for i, item := range result.Kvs {
		objIDList[i] = lastToken(string(item.Key))
	}
	return objIDList, nil
}

func (tracker *NodeTracker) EtcdGetObjectInfo(ctx context.Context, objIDList []string) ([]*pbNodeTracker.ObjectInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetObjectInfo")
	defer span.Finish()

	objInfoList := make([]*pbNodeTracker.ObjectInfo, len(objIDList))
	for i, objId := range objIDList {
		result, err := tracker.EtcdClient.Get(ctx, fmt.Sprintf("/object2hostname/%s/", objId), clientv3.WithPrefix())
		if err != nil {
			return nil, errors.Wrapf(err, "NodeTracker.EtcdGetObjectInfo objId: %v", objId)
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdObjectDelete")
	defer span.Finish()

	ops := make([]clientv3.Op, 0, len(objIDList))
	for _, objID := range objIDList {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/object2hostname/%s/", objID), clientv3.WithPrefix()))
	}
	_, err := tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		err = errors.Wrapf(err, "NodeTracker.EtcdObjectDelete")
	}
	return err
}

func (tracker *NodeTracker) EtcdGetRpcTarget(ctx context.Context, server string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.EtcdGetRpcTarget")
	defer span.Finish()

	result, err := tracker.EtcdClient.Get(ctx, fmt.Sprintf("/storage_server/%s/rpc_target", server))
	if err != nil {
		return "", errors.Wrapf(err, "NodeTracker.EtcdGetRpcTarget")
	}
	if len(result.Kvs) == 0 {
		err := errors.Errorf("RPC Target of %s not found", server)
		log.Error(err)
		return "", err
	}
	return string(result.Kvs[0].Value), nil
}

func (tracker *NodeTracker) ObjectInfo(ctx context.Context, in *pbNodeTracker.ObjectInfoRequest) (*pbNodeTracker.ObjectInfoResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.ObjectInfo")
	defer span.Finish()

	objIDList := in.GetObjectIdHexList()
	objInfoList, err := tracker.EtcdGetObjectInfo(ctx, objIDList)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &pbNodeTracker.ObjectInfoResponse{ObjectInfoList: objInfoList}, nil
}

func (tracker *NodeTracker) objectDeleteOnServer(ctx context.Context, server string, objIDList []string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.objectDeleteOnServer")
	defer span.Finish()
	// todo cache rpcTarget
	rpcTarget, err := tracker.EtcdGetRpcTarget(ctx, server)
	if err != nil {
		return errors.Wrapf(err, "EtcdGetRpcTarget server: %v", server)
	}
	span1, _ := opentracing.StartSpanFromContext(ctx, "grpc.Dial")
	conn, err := grpc.DialContext(ctx, rpcTarget, util.GrpcDialOption()...)
	if err != nil {
		span1.LogFields(ot_log.Error(err))
		span1.Finish()
		return errors.Wrapf(err, "grpc.DialContext, server:%v, rpc target: %v", server, rpcTarget)
	}
	span1.Finish()
	defer conn.Close()
	c := pbStorageServer.NewObjectStoreClient(conn)
	_, err = c.Delete(ctx, &pbStorageServer.DeleteRequest{ObjectIdHexList: objIDList})
	if err != nil {
		return errors.Wrapf(err, "rpc call Delete, object id list: %v", objIDList)
	}
	return nil
}

func (tracker *NodeTracker) objectDeleteDaemon(ref opentracing.SpanReference, objIDList []string) error {
	span := opentracing.StartSpan("objectDeleteDaemon", ref)
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	objInfoList, err := tracker.EtcdGetObjectInfo(ctx, objIDList)
	if err != nil {
		return errors.Wrapf(err, "objectDeleteDaemon")
	}
	serverObjMap := make(map[string][]string)
	for _, item := range objInfoList {
		for _, server := range item.ServerHostnameList {
			serverObjMap[server] = append(serverObjMap[server], item.ObjectIdHex)
		}
	}

	log.Debugf("serverObjMap to delete: %v", serverObjMap)

	resultQ := make(chan error, len(serverObjMap))
	for server, objList := range serverObjMap {
		go func(serverHostname string, objIDList []string) {
			err := tracker.objectDeleteOnServer(ctx, serverHostname, objIDList)
			resultQ <- errors.Wrapf(err, "NodeTracker.objectDeleteDaemon")
		}(server, objList)
	}

	for i := 0; i < len(serverObjMap); i++ {
		err := <-resultQ
		if err != nil {
			log.Errorf("%+v", err)
		}
	}

	var ops []clientv3.Op
	for _, objID := range objIDList {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/object2hostname/%s/", objID), clientv3.WithPrefix()))
	}

	for hostname, objIDList := range serverObjMap {
		for _, objID := range objIDList {
			key := fmt.Sprintf("/hostname2object/%s/%s", hostname, objID)
			ops = append(ops, clientv3.OpDelete(key))
		}

	}

	_, err = tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
	return errors.Wrapf(err, "NodeTracker.objectDeleteDaemon")
}

func (tracker *NodeTracker) ObjectDelete(ctx context.Context, in *pbNodeTracker.ObjectDeleteRequest) (*pbNodeTracker.ObjectDeleteResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.ObjectDelete")
	defer span.Finish()

	ref := opentracing.FollowsFrom(span.Context())

	objIDList := in.GetObjectIdHexList()
	tracker.deleteTaskQueue <- func() error {
		return tracker.objectDeleteDaemon(ref, objIDList)
	}

	return &pbNodeTracker.ObjectDeleteResponse{}, nil
}

func (tracker *NodeTracker) RegisterObject(ctx context.Context, in *pbNodeTracker.RegisterObjectRequest) (*pbNodeTracker.RegisterObjectResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.RegisterObject")
	defer span.Finish()

	objID := in.GetObjectIdHex()
	serverHostname := in.GetServerHostname()
	pushHostnameList := in.GetPushHostnameList()
	pushGroupList := in.GetPushGroupList()

	if len(pushGroupList) > 0 || len(pushHostnameList) > 0 {
		log.Debugf("pushHostnameList: %v", pushHostnameList)
		log.Debugf("pushGroupList: %v", pushGroupList)
	}

	ops := []clientv3.Op{
		clientv3.OpPut(fmt.Sprintf("/object2hostname/%s/%s", objID, serverHostname), "0"),
		clientv3.OpPut(fmt.Sprintf("/hostname2object/%s/%s", serverHostname, objID), "0"),
	}
	{ // etcd ctx
		span, ctx := opentracing.StartSpanFromContext(ctx, "etcd.client.put")
		_, err := tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
		span.Finish()
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	if len(pushHostnameList) > 0 || len(pushGroupList) > 0 {
		span := opentracing.SpanFromContext(ctx)
		span = opentracing.StartSpan("fetch_task_group", opentracing.FollowsFrom(span.Context()))
		ctx := opentracing.ContextWithSpan(context.Background(), span)
		tracker.fetchTaskQueue <- func() error {
			return tracker.Fetch(ctx, objID, serverHostname, pushHostnameList, pushGroupList)
		}
	}

	return &pbNodeTracker.RegisterObjectResponse{}, nil
}

func (tracker *NodeTracker) RegisterStorageClient(ctx context.Context, in *pbNodeTracker.StorageClient) (*pbNodeTracker.RegisterStorageClientResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.RegisterStorageClient")
	defer span.Finish()

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

func (tracker *NodeTracker) UnregisterStorageServer(ctx context.Context, in *pbNodeTracker.StorageServer) (*pbNodeTracker.UnregisterStorageServerResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.UnregisterStorageServer")
	defer span.Finish()

	hostname := in.GetHostname()
	objList, err := tracker.EtcdGetObjectList(ctx, hostname)
	if err != nil {
		return nil, err
	}
	serverInfo, err := tracker.EtcdGetStorageServer(ctx, hostname)
	if err != nil {
		return nil, err
	}

	ref := opentracing.FollowsFrom(span.Context())
	tracker.pushInfoTaskQueue <- func() error {
		return tracker.pushServerInfoToOthers(ref, hostname, true)
	}

	var ops []clientv3.Op
	ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/storage_server/%s/", hostname), clientv3.WithPrefix()))
	ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/hostname2object/%s/", hostname), clientv3.WithPrefix()))
	for _, obj := range objList {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/object2hostname/%s/%s", obj, hostname)))
	}
	for _, group := range serverInfo.GetGroupList() {
		ops = append(ops, clientv3.OpDelete(fmt.Sprintf("/storage_server_group/%s/%s", group, hostname)))
	}

	not_finish := true
	var ops_to_commit []clientv3.Op
	for not_finish {
		if len(ops) > EtcdTxnLimit {
			ops_to_commit = ops[:EtcdTxnLimit]
			ops = ops[EtcdTxnLimit:]
		} else {
			ops_to_commit = ops
			not_finish = false
		}
		_, err = tracker.EtcdClient.Txn(ctx).Then(ops_to_commit...).Commit()
		if err != nil {
			return nil, err
		}
	}

	log.Infof("Unregister storage server %s", hostname)
	return &pbNodeTracker.UnregisterStorageServerResponse{}, nil
}

func (tracker *NodeTracker) pushServerInfoToOthers(ref opentracing.SpanReference, hostname string, remove bool) error {
	span := opentracing.StartSpan("pushServerInfoToOthers", ref)
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	serverList, err := tracker.EtcdGetAllStorageServer(ctx)
	if err != nil {
		return errors.Wrapf(err, "pushServerInfoToOthers")
	}
	var infoToSend *pbNodeTracker.StorageServer
	for _, info := range serverList {
		if info.GetHostname() == hostname {
			infoToSend = info
		}
	}
	if infoToSend == nil {
		if remove {
			infoToSend = &pbNodeTracker.StorageServer{Hostname: hostname}
		} else {
			err := errors.Errorf("can not find %v", hostname)
			log.Errorf("%+v", err)
			return err
		}
	}

	ch := make(chan error, len(serverList))
	for _, info := range serverList {
		if info.GetHostname() == hostname {
			continue
		}

		go func(targetInfo *pbNodeTracker.StorageServer) {
			rpcTarget := targetInfo.GetRpcTarget()
			span, _ := opentracing.StartSpanFromContext(ctx, "grpc.DialContext")
			conn, err := grpc.DialContext(ctx, rpcTarget, util.GrpcDialOption()...)
			if err != nil {
				span.LogFields(ot_log.Error(err))
				span.Finish()
				ch <- errors.Wrapf(err, "failed to dial %v, rpc target %v", targetInfo.GetHostname(), rpcTarget)
				return
			}
			span.Finish()
			defer conn.Close()
			client := pbStorageServer.NewObjectStoreClient(conn)
			_, err = client.UpdateStorageServer(ctx, &pbStorageServer.UpdateStorageServerRequest{ServerInfo: infoToSend, Remove: remove})
			ch <- errors.Wrapf(err, "failed to update storage server on %v, rpc target %v", targetInfo.GetHostname(), rpcTarget)
		}(info)
	}

	for i := 0; i < len(serverList)-1; i++ {
		err := <-ch
		if err != nil {
			log.Errorf("%+v", err)
		}
	}
	log.Debugf("finish push server info of %v to all nodes", hostname)
	return nil
}

func (tracker *NodeTracker) RegisterStorageGroup(
	ctx context.Context, in *pbNodeTracker.StorageServer) (*pbNodeTracker.RegisterStorageGroupResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.RegisterStorageGroup")
	defer span.Finish()
	tracker.m.Lock()
	defer tracker.m.Unlock()

	hostname := in.GetHostname()
	groupList := in.GetGroupList()
	groupListKey := fmt.Sprintf("/storage_server/%s/group_list", hostname)
	result, err := tracker.EtcdClient.Get(ctx, groupListKey)
	if err != nil {
		return nil, errors.Wrapf(err, "NodeTracker.RegisterStorageGroup")
	}
	if len(result.Kvs) == 0 {
		return nil, errors.Errorf("hostname not found: ", hostname)
	}
	groupList = util.Unique(strings.Split(string(result.Kvs[0].Value), ","), groupList)
	ops := make([]clientv3.Op, 0, len(groupList)+1)
	ops = append(ops, clientv3.OpPut(groupListKey, strings.Join(groupList, ",")))
	for _, group := range groupList {
		ops = append(ops, clientv3.OpPut(fmt.Sprintf("/storage_server_group/%s/%s", group, hostname), ""))
	}
	_, err = tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return nil, err
	}
	return &pbNodeTracker.RegisterStorageGroupResponse{}, nil
}

func (tracker *NodeTracker) RegisterStorageServer(
	ctx context.Context, in *pbNodeTracker.StorageServer) (*pbNodeTracker.RegisterStorageServerResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.RegisterStorageServer")
	defer span.Finish()

	storageServerIP := getIPFromCtx(ctx)
	if in.GetIpAddr() != "" {
		storageServerIP = in.GetIpAddr()
	}

	if storageServerIP == "::1" || storageServerIP == "[::]" || storageServerIP == "127.0.0.1" {
		storageServerIP = tracker.RpcHost
	}

	hostname := in.GetHostname()
	rpcPort := in.GetRpcPort()
	plasmaSocket := in.GetPlasmaSocket()
	objTransferPort := in.GetObjTransferPort()
	groupList := in.GetGroupList()
	keyPrefix := fmt.Sprint("/storage_server/", hostname)
	items := map[string]string{
		"state":               "up",
		"ip_addr":             storageServerIP,
		"plasma_socket":       plasmaSocket,
		"rpc_port":            fmt.Sprint(rpcPort),
		"rpc_target":          fmt.Sprint(storageServerIP, ":", rpcPort),
		"obj_transfer_port":   fmt.Sprint(objTransferPort),
		"obj_transfer_target": fmt.Sprint(storageServerIP, ":", objTransferPort),
		"group_list":          strings.Join(groupList, ","),
	}
	{ // etcd ctx
		span, ctx := opentracing.StartSpanFromContext(ctx, "etcd.put")
		ops := make([]clientv3.Op, 0, len(items))
		for k, v := range items {
			ops = append(ops, clientv3.OpPut(fmt.Sprint(keyPrefix, "/", k), v))
		}
		for _, group := range groupList {
			ops = append(ops, clientv3.OpPut(fmt.Sprintf("/storage_server_group/%s/%s", group, hostname), ""))
		}

		_, err := tracker.EtcdClient.Txn(ctx).Then(ops...).Commit()
		span.Finish()

		if err != nil {
			log.Error(err)
			return nil, err
		}
	}
	ref := opentracing.FollowsFrom(span.Context())
	go tracker.pushServerInfoToOthers(ref, hostname, false)

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
	log.Info("start node tracker, listening on ", lis.Addr())

	grpcServer := grpc.NewServer(util.GrpcServerOption()...)
	pbNodeTracker.RegisterNodeTrackerServer(grpcServer, tracker)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %+v", err)
	}
}
