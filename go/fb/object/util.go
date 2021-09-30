package object

import (
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func FetchTaskHeaderToBytes(oid []byte, objectLength int64, createTime int64, version int64, spanCtx []byte, err []byte) []byte {
	if oid == nil {
		log.Fatalf("%+v", errors.New("ObjectMetaToBytes: oid is nil"))
	}
	builder := flatbuffers.NewBuilder(0)

	oidOffset := builder.CreateByteVector(oid)
	spanCtxOffset := builder.CreateByteVector(spanCtx)
	errOffset := builder.CreateByteVector(err)

	ObjectMetaStart(builder)
	ObjectMetaAddOid(builder, oidOffset)
	ObjectMetaAddCreateTime(builder, createTime)
	ObjectMetaAddVersion(builder, version)
	ObjectMetaAddObjectLength(builder, objectLength)
	objectMeta := ObjectMetaEnd(builder)

	FetchTaskHeaderStart(builder)
	FetchTaskHeaderAddObjectMeta(builder, objectMeta)
	FetchTaskHeaderAddSpanContext(builder, spanCtxOffset)
	FetchTaskHeaderAddError(builder, errOffset)
	header := FetchTaskHeaderEnd(builder)

	builder.Finish(header)
	return builder.FinishedBytes()
}

func BytesToFetchTaskHeader(buf []byte) *FetchTaskHeader {
	return GetRootAsFetchTaskHeader(buf, 0)
}
