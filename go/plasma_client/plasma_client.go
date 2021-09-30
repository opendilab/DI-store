package plasma_client

/*
#cgo linux LDFLAGS: -L${SRCDIR} -lplasma-c -lstdc++ -pthread
#cgo darwin LDFLAGS: -L${SRCDIR}/lib_darwin -lplasma -larrow -lplasma-cwrap -ljemalloc -lstdc++
#include <stdbool.h>
#include "cclient.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"context"
	"di_store/util"
	"encoding/hex"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

const (
	OBJECT_ID_LENGTH = 20
)

var once sync.Once

type Buff struct {
	inner  *C.Buff
	client *PlasmaClient
}

func NewBuff(buf *C.Buff, client *PlasmaClient) *Buff {
	return &Buff{buf, client}
}

func (buff *Buff) Data() unsafe.Pointer {
	if buff == nil {
		return nil
	}
	return unsafe.Pointer(buff.inner.data)
}

func (buff *Buff) Size() int {
	if buff == nil {
		return 0
	}
	return int(buff.inner.size)
}

func (buff *Buff) IsEmpty() bool {
	if buff == nil {
		return true
	}
	return unsafe.Pointer(buff.inner.data) == C.NULL
}

func (buff *Buff) Release(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Buff.Release")
	defer span.Finish()

	if buff == nil {
		return
	}

	buff.client.mu.Lock()
	defer buff.client.mu.Unlock()

	C.DeleteBuff(buff.inner)
}

func (buff *Buff) ToByteSlice() []byte {
	return util.BytesWithoutCopy(buff.Data(), buff.Size())
}

type PlasmaClient struct {
	mu      sync.Mutex
	cclient C.ClientPointer
}

func clientFinalizer(c *PlasmaClient) {
	if c.cclient != nil {
		c.Disconnect()
	}
}

func NewClient(path string) (*PlasmaClient, error) {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
	})
	c := &PlasmaClient{}
	runtime.SetFinalizer(c, clientFinalizer)
	err := c.Connect(path)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func msgToError(cmsg *C.char) error {
	msg := C.GoString(cmsg)
	if msg == "" {
		return nil
	} else {
		return errors.Errorf(msg)
	}
}

func (c *PlasmaClient) Connect(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cclient != nil {
		return errors.Errorf("client has connected to the server")
	}
	c.cclient = C.NewClient()
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	msg := C.Connect(c.cclient, cpath)
	return msgToError(msg)
}

func (c *PlasmaClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cclient == nil {
		return errors.Errorf("client has already been disconnected")
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

func (c *PlasmaClient) Create(ctx context.Context, oid []byte, length int) (*Buff, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Create")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(oid) != OBJECT_ID_LENGTH {
		panic(errors.Errorf("object id length error: %v", oid))
	}
	poid := unsafe.Pointer(&oid[0])

	var pbuff *C.Buff
	msg := C.Create(c.cclient, (*C.char)(poid), C.int(length), &pbuff)
	err := msgToError(msg)
	if err != nil {
		return nil, errors.Wrapf(err, "PlasmaClient.Create: %v", hex.EncodeToString(oid))
	}
	return NewBuff(pbuff, c), nil
}

func (c *PlasmaClient) Abort(ctx context.Context, oid []byte) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Abort")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(oid) != OBJECT_ID_LENGTH {
		panic(errors.Errorf("object id length error: %v", string(oid)))
	}
	poid := unsafe.Pointer(&oid[0])
	C.Abort(c.cclient, (*C.char)(poid))
}

func (c *PlasmaClient) Seal(ctx context.Context, oid []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Seal")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(oid) != OBJECT_ID_LENGTH {
		panic(errors.Errorf("object id length error: %v", string(oid)))
	}
	poid := unsafe.Pointer(&oid[0])

	msg := C.Seal(c.cclient, (*C.char)(poid))
	err := msgToError(msg)
	return err
}

func (c *PlasmaClient) Put(ctx context.Context, oid []byte, data []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Put")
	defer span.Finish()

	if len(oid) == 0 {
		oid = make([]byte, OBJECT_ID_LENGTH)
		rand.Read(oid)
	} else if len(oid) != OBJECT_ID_LENGTH {
		return errors.Errorf("object id length error: %v", string(oid))
	}

	buff, err := c.Create(ctx, oid, len(data))
	defer buff.Release(ctx)
	if err != nil {
		return err
	}

	{
		span, _ := opentracing.StartSpanFromContext(ctx, "PlasmaClient.ParallelMemCopy")
		pdata := unsafe.Pointer(&data[0])
		C.ParallelMemCopy2(buff.Data(), pdata, CInt64(buff.Size()), 1024*4, 8)
		span.Finish()
	}

	return c.Seal(ctx, oid)
}

func (c *PlasmaClient) Contains(ctx context.Context, oid []byte) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Contains")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(oid) != OBJECT_ID_LENGTH {
		return false, errors.Errorf("object id length error: %v", string(oid))
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

func (c *PlasmaClient) GetBuff(ctx context.Context, oid []byte) (*Buff, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.GetBuff")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	var pbuff *C.Buff
	msg := C.Get(c.cclient, (*C.char)(unsafe.Pointer(&oid[0])), CInt64(0), &pbuff)
	err := msgToError(msg)
	return NewBuff(pbuff, c), err
}

func (c *PlasmaClient) Get(ctx context.Context, oid []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Get")
	defer span.Finish()

	if len(oid) != OBJECT_ID_LENGTH {
		return nil, errors.Errorf("object id length error: %v", string(oid))
	}

	buff, err := c.GetBuff(ctx, oid)
	defer buff.Release(ctx)
	if err != nil {
		return nil, err
	}

	if !buff.IsEmpty() {

		span1, _ := opentracing.StartSpanFromContext(ctx, "PlasmaClient.makebyte")
		data := make([]byte, buff.Size(), buff.Size())
		span1.Finish()

		span2, _ := opentracing.StartSpanFromContext(ctx, "PlasmaClient.ParallelMemCopy")
		C.ParallelMemCopy2(unsafe.Pointer(&data[0]), buff.Data(), CInt64(buff.Size()), 1024*4, 8)
		span2.Finish()

		return data, nil
	} else {
		return nil, nil
	}
}

func (c *PlasmaClient) Delete(ctx context.Context, oids ...[]byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "PlasmaClient.Delete")
	defer span.Finish()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(oids) == 0 {
		return nil
	}
	objectIDs := make([]*C.char, len(oids))
	for i, oid := range oids {
		if len(oid) != OBJECT_ID_LENGTH {
			return errors.Errorf("object id length error: %v", string(oid))
		}
		poid := C.CBytes(oid)
		defer C.free(poid)
		objectIDs[i] = (*C.char)(poid)
	}
	pObjectIDs := (**C.char)(&objectIDs[0])
	msg := C.Delete(c.cclient, pObjectIDs, CInt64(len(oids)))
	return msgToError(msg)
}
