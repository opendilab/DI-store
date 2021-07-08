package plasma_client

// #cgo LDFLAGS: -L${SRCDIR}/libs -lplasma-cwrap -lplasma -larrow -ljemalloc -lstdc++

/*
#cgo LDFLAGS: -L${SRCDIR} -lplasma-c -lstdc++
#include <stdbool.h>
#include "cclient.h"
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/opentracing/opentracing-go"
)

const (
	OBJECT_ID_LENGTH = 20
)

var once sync.Once

type PlasmaClient struct {
	cclient C.ClientPointer
}

func clientFinalizer(c *PlasmaClient) {
	if c.cclient != nil {
		c.Disconnect()
	}
}

func NewClient() *PlasmaClient {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})
	c := &PlasmaClient{}
	runtime.SetFinalizer(c, clientFinalizer)
	return c
}

func msgToError(cmsg *C.char) error {
	msg := C.GoString(cmsg)
	if msg == "" {
		return nil
	} else {
		return fmt.Errorf(msg)
	}
}

func (c *PlasmaClient) Connect(path string) error {
	if c.cclient != nil {
		return fmt.Errorf("client has connected to the server")
	}
	c.cclient = C.NewClient()
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	msg := C.Connect(c.cclient, cpath)
	return msgToError(msg)
}

func (c *PlasmaClient) Disconnect() error {
	if c.cclient == nil {
		return fmt.Errorf("client has already been disconnected")
	}
	msg := C.Disconnect(c.cclient)
	C.DeleteClient(c.cclient)
	c.cclient = nil
	return msgToError(msg)
}

var (
	memcopy_threads   int32 = 4
	memcopy_threshold int64 = 0
	memcopy_blocksize int64 = 0
)

func (c *PlasmaClient) Put(ctx context.Context, oid []byte, data []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Put")
	defer span.Finish()

	evict_if_full := false

	if len(oid) == 0 {
		oid = make([]byte, OBJECT_ID_LENGTH)
		rand.Read(oid)
	} else if len(oid) != OBJECT_ID_LENGTH {
		return nil, fmt.Errorf("object id length error: %v", oid)
	}
	poid := unsafe.Pointer(&oid[0])
	pdata := unsafe.Pointer(&data[0])

	msg := C.CreateAndSeal(c.cclient, (*C.char)(poid), (*C.char)(pdata), C.int(len(data)),
		C.bool(evict_if_full), C.int(memcopy_threads), C.long(memcopy_threshold), C.long(memcopy_blocksize))
	err := msgToError(msg)
	if err != nil {
		return nil, err
	} else {
		return oid, nil
	}
}

func (c *PlasmaClient) Contains(ctx context.Context, oid []byte) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Contains")
	defer span.Finish()

	if len(oid) != OBJECT_ID_LENGTH {
		return false, fmt.Errorf("object id length error: %v", oid)
	}
	var hasObject C.char
	poid := unsafe.Pointer(&oid[0])
	msg := C.Contains(c.cclient, (*C.char)(poid), &hasObject)
	err := msgToError(msg)
	if err != nil {
		return false, err
	} else {
		return hasObject > 0, nil
	}
}

func (c *PlasmaClient) Get(ctx context.Context, oids ...[]byte) ([][]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Get")
	defer span.Finish()

	objectIDs := make([]*C.char, len(oids))
	for i, oid := range oids {
		if len(oid) != OBJECT_ID_LENGTH {
			return nil, fmt.Errorf("object id length error: %v", oid)
		}
		poid := C.CBytes(oid)
		defer C.free(poid)
		objectIDs[i] = (*C.char)(poid)
	}
	pObjectIDs := (**C.char)(&objectIDs[0])
	var buff C.ObjectBufferPointer
	pBuff := (*C.ObjectBufferPointer)(&buff)
	msg := C.Get(c.cclient, pObjectIDs, C.long(len(oids)), C.long(0), pBuff)
	defer C.DeleteObjectBufferPointer(buff)
	err := msgToError(msg)
	if err != nil {
		return nil, err
	}
	data := make([][]byte, len(oids))
	for i := 0; i < len(oids); i++ {
		item := C.GetData(buff, C.uint(i))
		pData := unsafe.Pointer(item.data)
		if pData != C.NULL {
			data[i] = C.GoBytes(pData, C.int(item.size))
		} else {
			data[i] = nil
		}
	}
	return data, nil
}
func (c *PlasmaClient) Delete(ctx context.Context, oids ...[]byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Delete")
	defer span.Finish()

	if len(oids) == 0 {
		return nil
	}
	objectIDs := make([]*C.char, len(oids))
	for i, oid := range oids {
		if len(oid) != OBJECT_ID_LENGTH {
			return fmt.Errorf("object id length error: %v", oid)
		}
		poid := C.CBytes(oid)
		defer C.free(poid)
		objectIDs[i] = (*C.char)(poid)
	}
	pObjectIDs := (**C.char)(&objectIDs[0])
	msg := C.Delete(c.cclient, pObjectIDs, C.long(len(oids)))
	return msgToError(msg)
}

type OpType int

const (
	OpPut OpType = iota
	OpGet
	OpDelete
	OpContains
)

type Task struct {
	Type    OpType
	Args    []interface{}
	ResultQ chan *OpResult
	Ctx     context.Context
}

type OpResult struct {
	Result interface{}
	Err    error
}

type PlasmaClientManager struct {
	TaskQ   chan *Task
	Clients []*PlasmaClient
}

func PlasmaClientWorker(taskQ <-chan *Task, client *PlasmaClient, id int) {
	for task := range taskQ {
		ctx := task.Ctx
		switch task.Type {
		case OpPut:
			oid := task.Args[0].([]byte)
			data := task.Args[1].([]byte)
			result, err := client.Put(ctx, oid, data)
			task.ResultQ <- &OpResult{Result: result, Err: err}
		case OpGet:
			oids := task.Args[0].([][]byte)
			result, err := client.Get(ctx, oids...)
			task.ResultQ <- &OpResult{Result: result, Err: err}
		case OpDelete:
			oids := task.Args[0].([][]byte)
			err := client.Delete(ctx, oids...)
			task.ResultQ <- &OpResult{Result: nil, Err: err}
		case OpContains:
			oid := task.Args[0].([]byte)
			hasObj, err := client.Contains(ctx, oid)
			task.ResultQ <- &OpResult{Result: hasObj, Err: err}
		}
	}
}

func NewPlasmaClientManager(n int, socketPath string) (*PlasmaClientManager, error) {
	clients := make([]*PlasmaClient, n)
	for i := range clients {
		client := NewClient()
		err := client.Connect(socketPath)
		if err != nil {
			return nil, err
		}
		clients[i] = client
	}
	taskQ := make(chan *Task, n)

	for i, client := range clients {
		go PlasmaClientWorker(taskQ, client, i)
	}
	return &PlasmaClientManager{TaskQ: taskQ, Clients: clients}, nil
}

func (m *PlasmaClientManager) Put(ctx context.Context, oid []byte, data []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClientManager.Put")
	defer span.Finish()

	resultQ := make(chan *OpResult, 1)
	task := &Task{
		Type:    OpPut,
		Args:    []interface{}{oid, data},
		ResultQ: resultQ,
		Ctx:     ctx,
	}
	m.TaskQ <- task
	result := <-resultQ
	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Result.([]byte), nil
	}
}

func (m *PlasmaClientManager) Contains(ctx context.Context, oid []byte) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClientManager.Contains")
	defer span.Finish()

	resultQ := make(chan *OpResult, 1)
	task := &Task{
		Type:    OpContains,
		Args:    []interface{}{oid},
		ResultQ: resultQ,
		Ctx:     ctx,
	}
	m.TaskQ <- task
	result := <-resultQ
	if result.Err != nil {
		return false, result.Err
	} else {
		return result.Result.(bool), nil
	}
}

func (m *PlasmaClientManager) Get(ctx context.Context, oids ...[]byte) ([][]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClientManager.Get")
	span.Finish()

	resultQ := make(chan *OpResult, 1)
	task := &Task{
		Type:    OpGet,
		Args:    []interface{}{oids},
		ResultQ: resultQ,
		Ctx:     ctx,
	}
	m.TaskQ <- task
	result := <-resultQ
	if result.Err != nil {
		return nil, result.Err
	} else {
		return result.Result.([][]byte), nil
	}
}

func (m *PlasmaClientManager) Delete(ctx context.Context, oids ...[]byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClientManager.Delete")
	defer span.Finish()

	resultQ := make(chan *OpResult, 1)
	task := &Task{
		Type:    OpDelete,
		Args:    []interface{}{oids},
		ResultQ: resultQ,
		Ctx:     ctx,
	}
	m.TaskQ <- task
	result := <-resultQ
	return result.Err
}
