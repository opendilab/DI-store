import grpc

from . import node_tracker_pb2
from . import node_tracker_pb2_grpc
from ..tracing import trace, wrap_channel

# todo grpc with multiprocessing
# https://github.com/grpc/grpc/issues/18321


@trace
class NodeTrackerClient:
    def __init__(self, node_tracker_host='127.0.0.1', node_tracker_port=50050):
        self.node_tracker_rpc_target = f'{node_tracker_host}:{node_tracker_port}'

        self.closed = False

    def close(self):
        if not self.closed:
            self.closed = True
            self.channel.close()

    def register_storage_server(self,
                                hostname,
                                ip_addr='', rpc_port=50051,
                                plasma_socket='/tmp/plasma'):
        assert hostname is not None
        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(channel)
            response = stub.register_storage_server(
                node_tracker_pb2.StorageServer(
                    hostname=hostname,
                    ip_addr=ip_addr,
                    rpc_port=rpc_port,
                    plasma_socket=plasma_socket
                ))
            return response

    def register_storage_client(self, server_hostname):
        assert server_hostname is not None
        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(wrap_channel(channel))
            response = stub.register_storage_client(
                node_tracker_pb2.StorageClient(
                    server_hostname=server_hostname
                ))
            return response

    def register_object(self, object_id_hex, server_hostname, push_hostname_list, push_group_list):
        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(wrap_channel(channel))
            response = stub.register_object(
                node_tracker_pb2.RegisterObjectRequest(
                    object_id_hex=object_id_hex,
                    server_hostname=server_hostname,
                    push_hostname_list=push_hostname_list,
                    push_group_list=push_group_list
                ))
            return response

    def server_info(self, server_hostname_list=None):
        request = node_tracker_pb2.ServerInfoRequest()
        if server_hostname_list:
            if isinstance(server_hostname_list, list):
                request.server_hostname_list.extend(server_hostname_list)
            else:
                request.server_hostname_list.append(server_hostname_list)
        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(wrap_channel(channel))
            response = stub.server_info(request)
            return response.storage_server_list

    def object_info(self, object_id_hex_list):
        request = node_tracker_pb2.ObjectInfoRequest()
        if isinstance(object_id_hex_list, list):
            request.object_id_hex_list.extend(object_id_hex_list)
        else:
            request.object_id_hex_list.append(object_id_hex_list)

        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(wrap_channel(channel))
            response = stub.object_info(request)
            return response.object_info_list

    def object_delelte(self, object_id_hex_list):
        request = node_tracker_pb2.ObjectDeleteRequest()
        if isinstance(object_id_hex_list, list):
            request.object_id_hex_list.extend(object_id_hex_list)
        else:
            request.object_id_hex_list.append(object_id_hex_list)

        with grpc.insecure_channel(self.node_tracker_rpc_target) as channel:
            stub = node_tracker_pb2_grpc.NodeTrackerStub(wrap_channel(channel))
            return stub.object_delete(request)
