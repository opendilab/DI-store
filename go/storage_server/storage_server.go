package storage_server

import (
	"context"
	"di_store/node_tracker"
	pbObjectStore "di_store/pb/storage_server"
	"di_store/plasma_client"
	plasma "di_store/plasma_server"
	"di_store/tracing"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

type StorageServer struct {
	pbObjectStore.UnimplementedObjectStoreServer
	Hostname           string
	PlasmaServerCtx    *context.Context
	PlasmaClient       *plasma_client.PlasmaClientManager
	NodeTrackerClient  *node_tracker.NodeTrackerClient
	Listener           *net.Listener
	ServerRpcTargetMap map[string]string
	mu                 sync.Mutex
	FetchRequestQ      chan *FetchTask
}

func NewStorageServer(
	ctx context.Context,
	hostname string,
	rpcPort int,
	nodeTrackerHost string,
	nodeTrackerPort int,
	plasmaSocket string,
	plasmaMemoryByte int,
	plasmaClientNum int,
	fetchWorkerNum int,
) (*StorageServer, error) {
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	}

	plasmaServerCtx, err := plasma.RunPlasmaStoreServer(ctx, plasmaMemoryByte, plasmaSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to run plasma store server: %v", err)
	}
	log.Info("running PlasmaStoreServer, path: ", plasmaSocket, " ,memory: ", plasmaMemoryByte)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", rpcPort))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	rpcPort = lis.Addr().(*net.TCPAddr).Port

	plasmaClient, err := plasma_client.NewPlasmaClientManager(plasmaClientNum, plasmaSocket)
	if err != nil {
		return nil, err
	}

	nodeTrackerClient, err := node_tracker.NewNodeTrackerClient(fmt.Sprintf("%s:%d", nodeTrackerHost, nodeTrackerPort))
	if err != nil {
		return nil, err
	}

	nodeTrackerClient.RegisterStorageServer(hostname, rpcPort, plasmaSocket)

	server := &StorageServer{
		Hostname:           hostname,
		PlasmaServerCtx:    &plasmaServerCtx,
		PlasmaClient:       plasmaClient,
		NodeTrackerClient:  nodeTrackerClient,
		Listener:           &lis,
		ServerRpcTargetMap: make(map[string]string),
		FetchRequestQ:      make(chan *FetchTask, fetchWorkerNum),
	}

	err = server.UpdateServerRpcTargetMap(context.Background())
	if err != nil {
		return nil, err
	}

	workerQ := make(chan *FetchTask, fetchWorkerNum)
	workerResultQ := make(chan *FetchTask, fetchWorkerNum)
	for i := 0; i < fetchWorkerNum; i++ {
		go fetchWorker(i, workerQ, workerResultQ, server)
	}

	go fetchWorkerManager(server.FetchRequestQ, workerQ, workerResultQ, server.PlasmaClient)

	return server, nil
}

func (server *StorageServer) UpdateServerRpcTargetMap(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.UpdateServerRpcTargetMap")
	defer span.Finish()
	server.mu.Lock()
	defer server.mu.Unlock()
	serverInfo, err := server.NodeTrackerClient.ServerInfo(ctx)
	if err != nil {
		return fmt.Errorf("UpdateServerRpcTargetMap: %v", err)
	}
	for _, info := range serverInfo {
		server.ServerRpcTargetMap[info.Hostname] = info.RpcTarget
	}
	return nil
}

func (server *StorageServer) GetRpcTargetOfServer(ctx context.Context, hostname string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "StorageServer.GetRpcTargetOfServer")
	defer span.Finish()
	rpcTarget, exist := server.ServerRpcTargetMap[hostname]
	if !exist {
		err := server.UpdateServerRpcTargetMap(ctx)
		if err != nil {
			return "", err
		}
		rpcTarget, exist = server.ServerRpcTargetMap[hostname]
	}
	if !exist {
		return "", fmt.Errorf("can not find rpc target of serer %s", hostname)
	}
	return rpcTarget, nil
}

func (server *StorageServer) Get(ctx context.Context, in *pbObjectStore.GetRequest) (*pbObjectStore.GetResponse, error) {
	oidHex := in.GetObjectIdHex()
	oid, err := hex.DecodeString(oidHex)
	if err != nil {
		return nil, err
	}
	dataList, err := server.PlasmaClient.Get(ctx, oid)
	if err != nil {
		return nil, err
	}
	data := dataList[0]
	// span := opentracing.SpanFromContext(ctx)

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
	log.Trace("Delete: ", oidHexList)
	return &pbObjectStore.DeleteResponse{}, nil
}

type FetchTask struct {
	Oid     []byte
	ResultQ chan error
	Ctx     context.Context
}

func fetchDataFromRemote(ctx context.Context, rpcTarget string, oid []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "fetchDataFromRemote")
	defer span.Finish()
	if log.IsLevelEnabled(log.TraceLevel) {
		t := time.Now()
		defer func() {
			d := time.Since(t)
			log.Trace(fmt.Sprintf("fetchDataFromRemote %s %s, duration: %s", rpcTarget, hex.EncodeToString(oid), d))
		}()
	}
	var client pbObjectStore.ObjectStoreClient
	{
		span, _ := opentracing.StartSpanFromContext(ctx, "fetchDataFromRemote.NewObjectStoreClient")

		conn, err := grpc.Dial(rpcTarget,
			append(tracing.GrpcDialOption,
				grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)),
				grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*2))...,
		)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client = pbObjectStore.NewObjectStoreClient(conn)
		span.Finish()
	}
	resp, err := client.Get(ctx, &pbObjectStore.GetRequest{ObjectIdHex: hex.EncodeToString(oid)})
	if err != nil {
		return nil, err
	}
	if resp.GetNotFound() {
		return nil, nil
	} else {
		return resp.GetData(), nil
	}
}

func fetchWorker(workerID int, workerQ <-chan *FetchTask, workerResultQ chan<- *FetchTask, server *StorageServer) {
	for task := range workerQ {
		ctx := task.Ctx

		if tracing.Enabled {
			span := opentracing.SpanFromContext(ctx)
			span.SetTag("worker_id", workerID)
		}

		objInfoList, err := server.NodeTrackerClient.ObjectInfo(ctx, hex.EncodeToString(task.Oid))

		if err != nil {
			task.ResultQ <- err
			workerResultQ <- task
			continue
		}
		objInfo := objInfoList[0]
		serverList := objInfo.ServerHostnameList
		rand.Shuffle(len(serverList), func(i, j int) {
			serverList[i], serverList[j] = serverList[j], serverList[i]
		})
		err = nil
		for _, s := range serverList {
			var rpcTarget string
			rpcTarget, err = server.GetRpcTargetOfServer(ctx, s)
			if err != nil {
				continue
			}
			var data []byte
			data, err = fetchDataFromRemote(ctx, rpcTarget, task.Oid)
			if err != nil {
				continue
			}
			if data != nil {
				_, err = server.PlasmaClient.Put(ctx, task.Oid, data)
				if err != nil {
					break
				}
				_, err = server.NodeTrackerClient.RegisterObject(ctx, hex.EncodeToString(task.Oid), server.Hostname)
				break
			}
		}
		task.ResultQ <- err
		workerResultQ <- task
	}
}

func fetchWorkerManager(requestQ <-chan *FetchTask, workerQ chan<- *FetchTask, workerResultQ <-chan *FetchTask, plasmaClient *plasma_client.PlasmaClientManager) {
	requestMap := make(map[string][]chan<- error)
	for {
		select {
		case task := <-workerResultQ:
			result := <-task.ResultQ
			key := string(task.Oid)
			for _, q := range requestMap[key] {
				q <- result
			}
			delete(requestMap, key)

		case task := <-requestQ:
			ctx := task.Ctx
			oid := task.Oid
			exist, err := plasmaClient.Contains(ctx, oid)
			if err != nil {
				task.ResultQ <- err
				continue
			}
			if exist {
				task.ResultQ <- nil
				continue
			}

			key := string(oid)
			l, exist := requestMap[key]

			if tracing.Enabled {
				span := opentracing.SpanFromContext(ctx)
				span.SetTag("request_exist", exist)
			}

			if !exist {
				workerQ <- &FetchTask{Oid: task.Oid, ResultQ: make(chan error, 1), Ctx: ctx}
			} else {
				log.Info(fmt.Sprintf("request exists: %s", hex.EncodeToString(task.Oid)))
			}
			requestMap[key] = append(l, task.ResultQ)
		}
	}

}

func (server *StorageServer) Fetch(ctx context.Context, in *pbObjectStore.FetchRequest) (*pbObjectStore.FetchResponse, error) {
	oidHex := in.GetObjectIdHex()
	oid, err := hex.DecodeString(oidHex)
	if err != nil {
		return nil, err
	}
	task := &FetchTask{Oid: oid, ResultQ: make(chan error, 1), Ctx: ctx}
	server.FetchRequestQ <- task
	err = <-task.ResultQ
	if err != nil {
		return nil, err
	}
	return &pbObjectStore.FetchResponse{}, nil
}

func (server *StorageServer) Serve() {

	log.Info("start storage server, listening on port ", (*server.Listener).Addr())
	grpcServer := grpc.NewServer(tracing.GrpcServerOption...)
	pbObjectStore.RegisterObjectStoreServer(grpcServer, server)
	if err := grpcServer.Serve(*server.Listener); err != nil {
		log.Fatalf("failed to create StorageServer: %v", err)
	}
}
