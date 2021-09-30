package metadata

import (
	"context"

	"github.com/opentracing/opentracing-go"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	clientv3.KV
}

func NewClient(client clientv3.KV) *EtcdClient {
	return &EtcdClient{client}
}

func (client *EtcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Put")
	defer span.Finish()

	return client.KV.Put(ctx, key, val, opts...)
}

func (client *EtcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Get")
	defer span.Finish()

	return client.KV.Get(ctx, key, opts...)
}

func (client *EtcdClient) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Delete")
	defer span.Finish()

	return client.KV.Delete(ctx, key, opts...)
}

func (client *EtcdClient) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Compact")
	defer span.Finish()

	return client.KV.Compact(ctx, rev, opts...)
}

func (client *EtcdClient) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Do")
	defer span.Finish()

	return client.KV.Do(ctx, op)
}

func (client *EtcdClient) Txn(ctx context.Context) clientv3.Txn {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EtcdClient.Txn")
	defer span.Finish()

	return &Txn{client.KV.Txn(ctx), ctx}
}

type Txn struct {
	clientv3.Txn
	ctx context.Context
}

func (txn *Txn) Commit() (*clientv3.TxnResponse, error) {
	span, _ := opentracing.StartSpanFromContext(txn.ctx, "Txn.Commit")
	defer span.Finish()

	return txn.Txn.Commit()
}

func (txn *Txn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &Txn{txn.Txn.If(cs...), txn.ctx}
}

func (txn *Txn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &Txn{txn.Txn.Then(ops...), txn.ctx}

}

func (txn *Txn) Else(ops ...clientv3.Op) clientv3.Txn {
	return &Txn{txn.Txn.Else(ops...), txn.ctx}
}
