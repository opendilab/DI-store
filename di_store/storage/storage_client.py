import socket
import sys
import time
import os
import threading
import numpy as np
import grpc
import pyarrow.plasma as plasma
from ..node_tracker.node_tracker_client import NodeTrackerClient
from . import storage_server_pb2
from . import storage_server_pb2_grpc

from ..common.config import Config
from ..tracing import trace, wrap_channel


def check_object_id_hex(object_id_hex):
    assert isinstance(object_id_hex, str) and len(object_id_hex) == 40


@trace
class StorageClient(object):
    def __init__(self, conf_path, hostname=None):
        config = Config(conf_path)
        node_tracker_info = config.node_tracker()[0]
        node_tracker_host = node_tracker_info['rpc_host']
        node_tracker_port = node_tracker_info['rpc_port']
        if hostname is None:
            hostname = os.getenv('DI_STORE_NODE_NAME')
        if hostname is None:
            hostname = socket.gethostname()
        self.hostname = hostname

        self.node_tracker_client = NodeTrackerClient(
            node_tracker_host, node_tracker_port)

        self.wait_for_server()

        assert self.server_rpc_target
        assert self.plasma_socket
        self.plasma_client = plasma.connect(self.plasma_socket)
        self.plasma_client_put = trace(self.plasma_client.put)
        self.plasma_client_get = trace(self.plasma_client.get)

        self.rng = np.random.default_rng()

        # todo grpc with multiprocessing
        # self.server_rpc_channel = wrap_channel(
        #     grpc.insecure_channel(self.server_rpc_target))

        # the expire-time set of objects with ttl
        self.expire_set = {}

    def wait_for_server(self):
        for i in range(1, 101):
            response = self.node_tracker_client.register_storage_client(
                self.hostname)
            if not response.storage_server.rpc_target:
                print(f'waiting for storage server {self.hostname}, retry {i}')
                time.sleep(1)
                continue
            else:
                self.server_rpc_target = response.storage_server.rpc_target
                self.plasma_socket = response.storage_server.plasma_socket
                self.etcd_server_ip_addr = response.etcd_server.ip_addr
                self.etcd_server_port = response.etcd_server.port
                break
        else:
            print(f'can not connect to storage server {self.hostname}')
            sys.exit(1)

    def register_group(self, group):
        if isinstance(group, list):
            group_list = group
        else:
            group_list = [group]
        self.node_tracker_client.register_group(self.hostname, group_list)

    @trace(span_name='span')
    def put(self, data, object_id=None, prefetch_hostname=None, prefetch_group=None, span=None):
        if object_id is not None:
            check_object_id_hex(object_id)
            object_id = plasma.ObjectID(bytes.fromhex(object_id))
        else:
            object_id = plasma.ObjectID(self.rng.bytes(20))

        if prefetch_hostname is None:
            prefetch_hostname = []
        elif not isinstance(prefetch_hostname, list):
            prefetch_hostname = [prefetch_hostname]

        if prefetch_group is None:
            prefetch_group = []
        elif not isinstance(prefetch_group, list):
            prefetch_group = [prefetch_group]

        ref = self.plasma_client_put(data, object_id)
        object_id_hex = ref.binary().hex()

        self.node_tracker_client.register_object(
            object_id_hex, self.hostname, prefetch_hostname, prefetch_group)
        if span:
            span.log_kv({'data_size': len(data)})
        return object_id_hex

    def fetch(self, object_id_hex, src_node=None, via_rpc=False):
        if src_node is None:
            src_node = ""
        with grpc.insecure_channel(self.server_rpc_target) as channel:
            channel = wrap_channel(channel)
            stub = storage_server_pb2_grpc.ObjectStoreStub(channel)
            stub.fetch(storage_server_pb2.FetchRequest(
                object_id_hex=object_id_hex, src_node=src_node, via_rpc=via_rpc))

    @trace(span_name='span')
    def get(self, object_id_hex, src_node=None, via_rpc=False, *, span=None):
        check_object_id_hex(object_id_hex)
        object_id = plasma.ObjectID(bytes.fromhex(object_id_hex))
        data = self.plasma_client_get(object_id, 0)
        if data != plasma.ObjectNotAvailable:
            if span:
                span.log_kv({'data_size': len(data)})
            return data

        self.fetch(object_id_hex, src_node, via_rpc)
        data = self.plasma_client_get(object_id, 0)
        if data == plasma.ObjectNotAvailable:
            if span:
                span.log_kv({'data_size': 0})
            return None
        else:
            if span:
                span.log_kv({'data_size': len(data)})
            return data

    def delete(self, object_id_hex_list):
        self.node_tracker_client.object_delelte(object_id_hex_list)

    # the following methods if the implement of TTL function¬
    @trace(span_name='span')
    def ttlGet(self, object_id_hex, src_node=None, via_rpc=False, *, span=None):
        check_object_id_hex(object_id_hex)

        # ttl process
        if object_id_hex in self.expire_set:
            end_time = self.expire_set[object_id_hex]
            current_time = time.time()
            if current_time > end_time:
                self.delete(object_id_hex)
                if span:
                    span.log_kv({'data_size': 0})
                return None

        object_id = plasma.ObjectID(bytes.fromhex(object_id_hex))
        data = self.plasma_client_get(object_id, 0)

        if data != plasma.ObjectNotAvailable:
            if span:
                span.log_kv({'data_size': len(data)})
            return data

        self.fetch(object_id_hex, src_node, via_rpc)
        data = self.plasma_client_get(object_id, 0)
        if data == plasma.ObjectNotAvailable:
            if span:
                span.log_kv({'data_size': 0})
            return None
        else:
            if span:
                span.log_kv({'data_size': len(data)})
            return data

    @trace(span_name='span')
    def ttlPut(self, data, object_id=None, prefetch_hostname=None, prefetch_group=None, span=None, ttl=0):
        if object_id is not None:
            check_object_id_hex(object_id)
            object_id = plasma.ObjectID(bytes.fromhex(object_id))
        else:
            object_id = plasma.ObjectID(self.rng.bytes(20))

        if prefetch_hostname is None:
            prefetch_hostname = []
        elif not isinstance(prefetch_hostname, list):
            prefetch_hostname = [prefetch_hostname]

        if prefetch_group is None:
            prefetch_group = []
        elif not isinstance(prefetch_group, list):
            prefetch_group = [prefetch_group]

        ref = self.plasma_client_put(data, object_id)
        object_id_hex = ref.binary().hex()

        self.node_tracker_client.register_object(
            object_id_hex, self.hostname, prefetch_hostname, prefetch_group)
        if span:
            span.log_kv({'data_size': len(data)})

        # ttl process. the time unit of ttl is second
        if ttl:
            self.removeTTLObjects()
            expire_time = self.getExpireTime(ttl)
            self.expire_set[object_id_hex] = expire_time

        return object_id_hex

    def getExpireTime(self, ttl=0):
        start_time = time.time()
        return start_time + ttl

    def removeTTLObjects(self):
        current_time = time.time()
        for key in list(self.expire_set.keys()):
            if self.expire_set[key] < current_time:
                self.delete(key)
                self.expire_set.pop(key)





thread_local = threading.local()


class Client:
    def __init__(self, conf_path, hostname=None):
        self.conf_path = conf_path
        self.hostname = hostname
        self._get_local_client()

    def _get_local_client(self):
        current_id = os.getpid(), threading.get_ident()
        client_id, client = getattr(
            thread_local,
            self.conf_path,
            (None, None))
        if client_id != current_id:
            client = StorageClient(self.conf_path, self.hostname)
            setattr(
                thread_local,
                self.conf_path,
                (current_id, client)
            )
        return client

    def register_group(self, *args, **kwargs):
        return self._get_local_client().register_group(*args, **kwargs)

    def put(self, *args, **kwargs):
        return self._get_local_client().ttlPut(*args, **kwargs)

    def get(self, *args, **kwargs):
        return self._get_local_client().ttlGet(*args, **kwargs)

    def delete(self, *args, **kwargs):
        return self._get_local_client().delete(*args, **kwargs)

    def getExpireSet(self):
        return self._get_local_client().expire_set
