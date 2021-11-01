package node_tracker

import (
	"context"
	"di_store/util"
	"fmt"
	"math/rand"
	"sync"

	pbObjectStore "di_store/pb/storage_server"

	"github.com/opentracing/opentracing-go"
	ot_log "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type FetchFn func(ctx context.Context, objID string, src string, dst string) error

type ObjectFetchWorker struct {
	ObjID             string
	Ctx               context.Context
	AvailableNodeList []*SrcNodeState
	WaitingNodeList   []*DstNodeState
	FinishedNodeList  []string
	FailedNodeList    []string
	InProgressCnt     int
	mu                sync.Mutex
	cancel            context.CancelFunc
}

type SrcNodeState struct {
	Hostname   string
	VirtualID  int
	FailedCnt  int
	SucceedCnt int
}

type DstNodeState struct {
	Hostname  string
	FailedCnt int
}

func taskMainLoop(taskQueue <-chan func() error) {
	taskInflight := 0
	BackgroundTaskLimit := util.CommonConfig.BackgroundTaskLimit
	ch := make(chan error, BackgroundTaskLimit)
	processError := func(err error) {
		if err != nil {
			log.Errorf("%+v", err)
		}
	}
	for {
		switch taskInflight {
		case 0:
			task := <-taskQueue
			go func() { ch <- task() }()
			taskInflight += 1
		case BackgroundTaskLimit:
			err := <-ch
			taskInflight -= 1
			processError(err)
		default:
			select {
			case task := <-taskQueue:
				go func() { ch <- task() }()
				taskInflight += 1
			case err := <-ch:
				taskInflight -= 1
				processError(err)
			}
		}
	}
}

func distinctNodeList(srcNode string, l []string) []string {
	m := make(map[string]bool)
	for _, v := range l {
		m[v] = true
	}
	result := make([]string, 0, len(m))
	for k := range m {
		if k != srcNode {
			result = append(result, k)
		}
	}
	return result
}

func (tracker *NodeTracker) Fetch(ctx context.Context, objID string, srcNode string, dstNodeList []string, dstGroupList []string) error {
	// close span of fetch_task_group
	spanParent := opentracing.SpanFromContext(ctx)
	defer spanParent.Finish()

	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.Fetch")
	defer span.Finish()

	for _, group := range dstGroupList {
		ctx, cancel := context.WithTimeout(ctx, util.CommonConfig.RequestTimeout)
		defer cancel()

		nodeList, err := tracker.EtcdGetHostnameListOfGroup(ctx, group)
		if err != nil {
			return errors.Wrapf(err, "NodeTracker.Fetch")
		}
		dstNodeList = append(dstNodeList, nodeList...)
	}
	dstNodeList = distinctNodeList(srcNode, dstNodeList)
	log.Debugf("final node list: %v", dstNodeList)

	worker := InitObjectFetchWorker(ctx, objID, []string{srcNode}, dstNodeList)
	worker.DispatchTask(tracker.DoFetch)
	<-worker.Ctx.Done()

	log.Debugf("NodeTracker.Fetch finish, objID %v, srcNode %v, dstNodeList %v, dstGroupList %v", objID, srcNode, dstNodeList, dstGroupList)
	log.Debugf("FinishedNodeList: %v, FailedNodeList: %v", worker.FinishedNodeList, worker.FailedNodeList)

	return nil
}

func (tracker *NodeTracker) DoFetch(ctx context.Context, objID string, src string, dst string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "NodeTracker.DoFetch")
	defer span.Finish()

	span.LogFields(
		ot_log.String("src", src),
		ot_log.String("dst", dst),
	)
	log.Debugf("fetch: %v -> %v", src, dst)
	dstNodeInfo, err := tracker.EtcdGetStorageServer(ctx, dst) // todo with cache
	if err != nil {
		return err
	}
	span1 := opentracing.StartSpan("grpc.Dial", opentracing.ChildOf(span.Context()))
	conn, err := grpc.DialContext(ctx, dstNodeInfo.GetRpcTarget(), util.GrpcDialOption()...)
	if err != nil {
		span1.LogFields(ot_log.Error(err))
		span1.Finish()
		return err
	}
	span1.Finish()
	defer conn.Close()

	client := pbObjectStore.NewObjectStoreClient(conn)
	req := pbObjectStore.FetchRequest{
		ObjectIdHex: objID,
		ViaRpc:      false,
		SrcNode:     src,
		SrcNodeOnly: true,
	}
	_, err = client.Fetch(ctx, &req)
	return err
}

func InitObjectFetchWorker(ctx context.Context, objID string, availableNodeList []string, waitingNodeList []string) *ObjectFetchWorker {
	ctx, cancel := context.WithCancel(ctx)
	span, ctx := opentracing.StartSpanFromContext(ctx, "InitObjectFetchWorker")
	defer span.Finish()

	virtualNode := util.CommonConfig.FetchSrcVirtualNodeNumber
	_availableNodeList := make([]*SrcNodeState, 0, len(availableNodeList)*virtualNode)
	for i := 0; i < len(availableNodeList); i++ {
		for virtualID := 0; virtualID < virtualNode; virtualID++ {
			_availableNodeList = append(
				_availableNodeList, &SrcNodeState{Hostname: availableNodeList[i], VirtualID: virtualID})
		}
	}

	_waitingNodeList := make([]*DstNodeState, len(waitingNodeList))
	for i := 0; i < len(waitingNodeList); i++ {
		_waitingNodeList[i] = &DstNodeState{Hostname: waitingNodeList[i]}
	}

	return &ObjectFetchWorker{
		ObjID:             objID,
		AvailableNodeList: _availableNodeList,
		WaitingNodeList:   _waitingNodeList,
		Ctx:               ctx,
		cancel:            cancel,
	}
}

func (w *ObjectFetchWorker) DispatchTask(fetch FetchFn) {
	span, _ := opentracing.StartSpanFromContext(w.Ctx, "ObjectFetchWorker.DispatchTask")
	defer span.Finish()

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.WaitingNodeList) == 0 && w.InProgressCnt == 0 {
		w.cancel()
		return
	}

	rand.Shuffle(len(w.AvailableNodeList), func(i, j int) {
		w.AvailableNodeList[i], w.AvailableNodeList[j] = w.AvailableNodeList[j], w.AvailableNodeList[i]
	})

	rand.Shuffle(len(w.WaitingNodeList), func(i, j int) {
		w.WaitingNodeList[i], w.WaitingNodeList[j] = w.WaitingNodeList[j], w.WaitingNodeList[i]
	})

	virtualNodeFanout := util.CommonConfig.FetchSrcVirtualNodeFanout
	retryMax := util.CommonConfig.FetchTaskRetryMax
	virtualNodeNumber := util.CommonConfig.FetchSrcVirtualNodeNumber
	log.Debugf("ObjectFetchWorker.DispatchTask, availableNodeList: %+v, waitingNodeList: %+v, VirtualNodeNumber: %+v, VirtualNodeFanout: %v, retryMax: %v",
		w.AvailableNodeList, w.WaitingNodeList, virtualNodeNumber, virtualNodeFanout, retryMax)

	n := min(len(w.AvailableNodeList), len(w.WaitingNodeList))
	for i := 0; i < n; i++ {
		src := w.AvailableNodeList[i]
		dst := w.WaitingNodeList[i]
		go func() {
			log.Debugf("fetch: %+v -> %+v", src, dst)
			err := fetch(w.Ctx, w.ObjID, src.Hostname, dst.Hostname)
			w.mu.Lock()
			defer w.mu.Unlock()
			if err != nil {
				src.FailedCnt += 1
				if src.FailedCnt < retryMax {
					w.AvailableNodeList = append(w.AvailableNodeList, src)
				} else {
					log.Warnf("fetch task: node %v failed %v time/times", src, retryMax)
				}

				dst.FailedCnt += 1
				if dst.FailedCnt < retryMax {
					w.WaitingNodeList = append(w.WaitingNodeList, dst)
				} else {
					log.Warnf("fetch task: node %v failed %v time/times", dst.Hostname, retryMax)
					w.FailedNodeList = append(w.FailedNodeList, dst.Hostname)
				}
				log.Errorf("%+v", err)
			} else {
				src.FailedCnt = 0
				dst.FailedCnt = 0
				src.SucceedCnt += 1
				if src.SucceedCnt < virtualNodeFanout {
					w.AvailableNodeList = append(w.AvailableNodeList, src)
				} else {
					log.Debugf("%v (%v) has reached the fetch fanout limit %v", src.Hostname, src.VirtualID, virtualNodeFanout)
				}
				for virtualID := 0; virtualID < virtualNodeNumber; virtualID++ {
					w.AvailableNodeList = append(w.AvailableNodeList, &SrcNodeState{
						Hostname:  dst.Hostname,
						VirtualID: virtualID,
					})
				}

				w.FinishedNodeList = append(w.FinishedNodeList, dst.Hostname)
			}
			w.InProgressCnt -= 1
			go w.DispatchTask(fetch)
		}()
	}
	w.AvailableNodeList = w.AvailableNodeList[n:]
	w.WaitingNodeList = w.WaitingNodeList[n:]
	w.InProgressCnt += n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *SrcNodeState) String() string {
	return fmt.Sprintf(
		"{Hostname: %v (%v), FailedCnt: %v, SucceedCnt: %v}",
		s.Hostname, s.VirtualID, s.FailedCnt, s.SucceedCnt)
}

func (s *DstNodeState) String() string {
	return fmt.Sprintf(
		"{Hostname: %v, FailedCnt: %v}",
		s.Hostname, s.FailedCnt)
}
