# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: node_tracker.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='node_tracker.proto',
  package='di_store.node_tracker',
  syntax='proto3',
  serialized_options=b'Z\030di_store/pb/node_tracker',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x12node_tracker.proto\x12\x15\x64i_store.node_tracker\"1\n\x13ObjectDeleteRequest\x12\x1a\n\x12object_id_hex_list\x18\x01 \x03(\t\"\x16\n\x14ObjectDeleteResponse\"/\n\x11ObjectInfoRequest\x12\x1a\n\x12object_id_hex_list\x18\x01 \x03(\t\"Q\n\x12ObjectInfoResponse\x12;\n\x10object_info_list\x18\x01 \x03(\x0b\x32!.di_store.node_tracker.ObjectInfo\"A\n\nObjectInfo\x12\x15\n\robject_id_hex\x18\x01 \x01(\t\x12\x1c\n\x14server_hostname_list\x18\x02 \x03(\t\"1\n\x11ServerInfoRequest\x12\x1c\n\x14server_hostname_list\x18\x01 \x03(\t\"W\n\x12ServerInfoResponse\x12\x41\n\x13storage_server_list\x18\x01 \x03(\x0b\x32$.di_store.node_tracker.StorageServer\"|\n\x15RegisterObjectRequest\x12\x15\n\robject_id_hex\x18\x01 \x01(\t\x12\x17\n\x0fserver_hostname\x18\x02 \x01(\t\x12\x1a\n\x12push_hostname_list\x18\x03 \x03(\t\x12\x17\n\x0fpush_group_list\x18\x04 \x03(\t\"\x18\n\x16RegisterObjectResponse\"\xbb\x01\n\rStorageServer\x12\x10\n\x08hostname\x18\x01 \x01(\t\x12\x0f\n\x07ip_addr\x18\x02 \x01(\t\x12\x15\n\rplasma_socket\x18\x03 \x01(\t\x12\x10\n\x08rpc_port\x18\x04 \x01(\x05\x12\x12\n\nrpc_target\x18\x05 \x01(\t\x12\x19\n\x11obj_transfer_port\x18\x06 \x01(\x05\x12\x1b\n\x13obj_transfer_target\x18\x07 \x01(\t\x12\x12\n\ngroup_list\x18\x08 \x03(\t\"w\n\x1dRegisterStorageServerResponse\x12\x36\n\x0b\x65tcd_server\x18\x01 \x01(\x0b\x32!.di_store.node_tracker.EtcdServer\x12\x1e\n\x16storage_server_ip_addr\x18\x02 \x01(\t\"(\n\rStorageClient\x12\x17\n\x0fserver_hostname\x18\x01 \x01(\t\"\x95\x01\n\x1dRegisterStorageClientResponse\x12\x36\n\x0b\x65tcd_server\x18\x01 \x01(\x0b\x32!.di_store.node_tracker.EtcdServer\x12<\n\x0estorage_server\x18\x02 \x01(\x0b\x32$.di_store.node_tracker.StorageServer\"\x1e\n\x1cRegisterStorageGroupResponse\" \n\x1eUnregisterStorageGroupResponse\"+\n\nEtcdServer\x12\x0f\n\x07ip_addr\x18\x01 \x01(\t\x12\x0c\n\x04port\x18\x02 \x01(\x05\"!\n\x1fUnregisterStorageServerResponse2\x98\x08\n\x0bNodeTracker\x12w\n\x17register_storage_server\x12$.di_store.node_tracker.StorageServer\x1a\x34.di_store.node_tracker.RegisterStorageServerResponse\"\x00\x12{\n\x19unregister_storage_server\x12$.di_store.node_tracker.StorageServer\x1a\x36.di_store.node_tracker.UnregisterStorageServerResponse\"\x00\x12u\n\x16register_storage_group\x12$.di_store.node_tracker.StorageServer\x1a\x33.di_store.node_tracker.RegisterStorageGroupResponse\"\x00\x12y\n\x18unregister_storage_group\x12$.di_store.node_tracker.StorageServer\x1a\x35.di_store.node_tracker.UnregisterStorageGroupResponse\"\x00\x12w\n\x17register_storage_client\x12$.di_store.node_tracker.StorageClient\x1a\x34.di_store.node_tracker.RegisterStorageClientResponse\"\x00\x12p\n\x0fregister_object\x12,.di_store.node_tracker.RegisterObjectRequest\x1a-.di_store.node_tracker.RegisterObjectResponse\"\x00\x12\x64\n\x0bserver_info\x12(.di_store.node_tracker.ServerInfoRequest\x1a).di_store.node_tracker.ServerInfoResponse\"\x00\x12\x64\n\x0bobject_info\x12(.di_store.node_tracker.ObjectInfoRequest\x1a).di_store.node_tracker.ObjectInfoResponse\"\x00\x12j\n\robject_delete\x12*.di_store.node_tracker.ObjectDeleteRequest\x1a+.di_store.node_tracker.ObjectDeleteResponse\"\x00\x42\x1aZ\x18\x64i_store/pb/node_trackerb\x06proto3'
)




_OBJECTDELETEREQUEST = _descriptor.Descriptor(
  name='ObjectDeleteRequest',
  full_name='di_store.node_tracker.ObjectDeleteRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='object_id_hex_list', full_name='di_store.node_tracker.ObjectDeleteRequest.object_id_hex_list', index=0,
      number=1, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=45,
  serialized_end=94,
)


_OBJECTDELETERESPONSE = _descriptor.Descriptor(
  name='ObjectDeleteResponse',
  full_name='di_store.node_tracker.ObjectDeleteResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=96,
  serialized_end=118,
)


_OBJECTINFOREQUEST = _descriptor.Descriptor(
  name='ObjectInfoRequest',
  full_name='di_store.node_tracker.ObjectInfoRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='object_id_hex_list', full_name='di_store.node_tracker.ObjectInfoRequest.object_id_hex_list', index=0,
      number=1, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=120,
  serialized_end=167,
)


_OBJECTINFORESPONSE = _descriptor.Descriptor(
  name='ObjectInfoResponse',
  full_name='di_store.node_tracker.ObjectInfoResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='object_info_list', full_name='di_store.node_tracker.ObjectInfoResponse.object_info_list', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=169,
  serialized_end=250,
)


_OBJECTINFO = _descriptor.Descriptor(
  name='ObjectInfo',
  full_name='di_store.node_tracker.ObjectInfo',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='object_id_hex', full_name='di_store.node_tracker.ObjectInfo.object_id_hex', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='server_hostname_list', full_name='di_store.node_tracker.ObjectInfo.server_hostname_list', index=1,
      number=2, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=252,
  serialized_end=317,
)


_SERVERINFOREQUEST = _descriptor.Descriptor(
  name='ServerInfoRequest',
  full_name='di_store.node_tracker.ServerInfoRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='server_hostname_list', full_name='di_store.node_tracker.ServerInfoRequest.server_hostname_list', index=0,
      number=1, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=319,
  serialized_end=368,
)


_SERVERINFORESPONSE = _descriptor.Descriptor(
  name='ServerInfoResponse',
  full_name='di_store.node_tracker.ServerInfoResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='storage_server_list', full_name='di_store.node_tracker.ServerInfoResponse.storage_server_list', index=0,
      number=1, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=370,
  serialized_end=457,
)


_REGISTEROBJECTREQUEST = _descriptor.Descriptor(
  name='RegisterObjectRequest',
  full_name='di_store.node_tracker.RegisterObjectRequest',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='object_id_hex', full_name='di_store.node_tracker.RegisterObjectRequest.object_id_hex', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='server_hostname', full_name='di_store.node_tracker.RegisterObjectRequest.server_hostname', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='push_hostname_list', full_name='di_store.node_tracker.RegisterObjectRequest.push_hostname_list', index=2,
      number=3, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='push_group_list', full_name='di_store.node_tracker.RegisterObjectRequest.push_group_list', index=3,
      number=4, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=459,
  serialized_end=583,
)


_REGISTEROBJECTRESPONSE = _descriptor.Descriptor(
  name='RegisterObjectResponse',
  full_name='di_store.node_tracker.RegisterObjectResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=585,
  serialized_end=609,
)


_STORAGESERVER = _descriptor.Descriptor(
  name='StorageServer',
  full_name='di_store.node_tracker.StorageServer',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='hostname', full_name='di_store.node_tracker.StorageServer.hostname', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='ip_addr', full_name='di_store.node_tracker.StorageServer.ip_addr', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='plasma_socket', full_name='di_store.node_tracker.StorageServer.plasma_socket', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='rpc_port', full_name='di_store.node_tracker.StorageServer.rpc_port', index=3,
      number=4, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='rpc_target', full_name='di_store.node_tracker.StorageServer.rpc_target', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='obj_transfer_port', full_name='di_store.node_tracker.StorageServer.obj_transfer_port', index=5,
      number=6, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='obj_transfer_target', full_name='di_store.node_tracker.StorageServer.obj_transfer_target', index=6,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='group_list', full_name='di_store.node_tracker.StorageServer.group_list', index=7,
      number=8, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=612,
  serialized_end=799,
)


_REGISTERSTORAGESERVERRESPONSE = _descriptor.Descriptor(
  name='RegisterStorageServerResponse',
  full_name='di_store.node_tracker.RegisterStorageServerResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='etcd_server', full_name='di_store.node_tracker.RegisterStorageServerResponse.etcd_server', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='storage_server_ip_addr', full_name='di_store.node_tracker.RegisterStorageServerResponse.storage_server_ip_addr', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=801,
  serialized_end=920,
)


_STORAGECLIENT = _descriptor.Descriptor(
  name='StorageClient',
  full_name='di_store.node_tracker.StorageClient',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='server_hostname', full_name='di_store.node_tracker.StorageClient.server_hostname', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=922,
  serialized_end=962,
)


_REGISTERSTORAGECLIENTRESPONSE = _descriptor.Descriptor(
  name='RegisterStorageClientResponse',
  full_name='di_store.node_tracker.RegisterStorageClientResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='etcd_server', full_name='di_store.node_tracker.RegisterStorageClientResponse.etcd_server', index=0,
      number=1, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='storage_server', full_name='di_store.node_tracker.RegisterStorageClientResponse.storage_server', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=965,
  serialized_end=1114,
)


_REGISTERSTORAGEGROUPRESPONSE = _descriptor.Descriptor(
  name='RegisterStorageGroupResponse',
  full_name='di_store.node_tracker.RegisterStorageGroupResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1116,
  serialized_end=1146,
)


_UNREGISTERSTORAGEGROUPRESPONSE = _descriptor.Descriptor(
  name='UnregisterStorageGroupResponse',
  full_name='di_store.node_tracker.UnregisterStorageGroupResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1148,
  serialized_end=1180,
)


_ETCDSERVER = _descriptor.Descriptor(
  name='EtcdServer',
  full_name='di_store.node_tracker.EtcdServer',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='ip_addr', full_name='di_store.node_tracker.EtcdServer.ip_addr', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='port', full_name='di_store.node_tracker.EtcdServer.port', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1182,
  serialized_end=1225,
)


_UNREGISTERSTORAGESERVERRESPONSE = _descriptor.Descriptor(
  name='UnregisterStorageServerResponse',
  full_name='di_store.node_tracker.UnregisterStorageServerResponse',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1227,
  serialized_end=1260,
)

_OBJECTINFORESPONSE.fields_by_name['object_info_list'].message_type = _OBJECTINFO
_SERVERINFORESPONSE.fields_by_name['storage_server_list'].message_type = _STORAGESERVER
_REGISTERSTORAGESERVERRESPONSE.fields_by_name['etcd_server'].message_type = _ETCDSERVER
_REGISTERSTORAGECLIENTRESPONSE.fields_by_name['etcd_server'].message_type = _ETCDSERVER
_REGISTERSTORAGECLIENTRESPONSE.fields_by_name['storage_server'].message_type = _STORAGESERVER
DESCRIPTOR.message_types_by_name['ObjectDeleteRequest'] = _OBJECTDELETEREQUEST
DESCRIPTOR.message_types_by_name['ObjectDeleteResponse'] = _OBJECTDELETERESPONSE
DESCRIPTOR.message_types_by_name['ObjectInfoRequest'] = _OBJECTINFOREQUEST
DESCRIPTOR.message_types_by_name['ObjectInfoResponse'] = _OBJECTINFORESPONSE
DESCRIPTOR.message_types_by_name['ObjectInfo'] = _OBJECTINFO
DESCRIPTOR.message_types_by_name['ServerInfoRequest'] = _SERVERINFOREQUEST
DESCRIPTOR.message_types_by_name['ServerInfoResponse'] = _SERVERINFORESPONSE
DESCRIPTOR.message_types_by_name['RegisterObjectRequest'] = _REGISTEROBJECTREQUEST
DESCRIPTOR.message_types_by_name['RegisterObjectResponse'] = _REGISTEROBJECTRESPONSE
DESCRIPTOR.message_types_by_name['StorageServer'] = _STORAGESERVER
DESCRIPTOR.message_types_by_name['RegisterStorageServerResponse'] = _REGISTERSTORAGESERVERRESPONSE
DESCRIPTOR.message_types_by_name['StorageClient'] = _STORAGECLIENT
DESCRIPTOR.message_types_by_name['RegisterStorageClientResponse'] = _REGISTERSTORAGECLIENTRESPONSE
DESCRIPTOR.message_types_by_name['RegisterStorageGroupResponse'] = _REGISTERSTORAGEGROUPRESPONSE
DESCRIPTOR.message_types_by_name['UnregisterStorageGroupResponse'] = _UNREGISTERSTORAGEGROUPRESPONSE
DESCRIPTOR.message_types_by_name['EtcdServer'] = _ETCDSERVER
DESCRIPTOR.message_types_by_name['UnregisterStorageServerResponse'] = _UNREGISTERSTORAGESERVERRESPONSE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ObjectDeleteRequest = _reflection.GeneratedProtocolMessageType('ObjectDeleteRequest', (_message.Message,), {
  'DESCRIPTOR' : _OBJECTDELETEREQUEST,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ObjectDeleteRequest)
  })
_sym_db.RegisterMessage(ObjectDeleteRequest)

ObjectDeleteResponse = _reflection.GeneratedProtocolMessageType('ObjectDeleteResponse', (_message.Message,), {
  'DESCRIPTOR' : _OBJECTDELETERESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ObjectDeleteResponse)
  })
_sym_db.RegisterMessage(ObjectDeleteResponse)

ObjectInfoRequest = _reflection.GeneratedProtocolMessageType('ObjectInfoRequest', (_message.Message,), {
  'DESCRIPTOR' : _OBJECTINFOREQUEST,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ObjectInfoRequest)
  })
_sym_db.RegisterMessage(ObjectInfoRequest)

ObjectInfoResponse = _reflection.GeneratedProtocolMessageType('ObjectInfoResponse', (_message.Message,), {
  'DESCRIPTOR' : _OBJECTINFORESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ObjectInfoResponse)
  })
_sym_db.RegisterMessage(ObjectInfoResponse)

ObjectInfo = _reflection.GeneratedProtocolMessageType('ObjectInfo', (_message.Message,), {
  'DESCRIPTOR' : _OBJECTINFO,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ObjectInfo)
  })
_sym_db.RegisterMessage(ObjectInfo)

ServerInfoRequest = _reflection.GeneratedProtocolMessageType('ServerInfoRequest', (_message.Message,), {
  'DESCRIPTOR' : _SERVERINFOREQUEST,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ServerInfoRequest)
  })
_sym_db.RegisterMessage(ServerInfoRequest)

ServerInfoResponse = _reflection.GeneratedProtocolMessageType('ServerInfoResponse', (_message.Message,), {
  'DESCRIPTOR' : _SERVERINFORESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.ServerInfoResponse)
  })
_sym_db.RegisterMessage(ServerInfoResponse)

RegisterObjectRequest = _reflection.GeneratedProtocolMessageType('RegisterObjectRequest', (_message.Message,), {
  'DESCRIPTOR' : _REGISTEROBJECTREQUEST,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.RegisterObjectRequest)
  })
_sym_db.RegisterMessage(RegisterObjectRequest)

RegisterObjectResponse = _reflection.GeneratedProtocolMessageType('RegisterObjectResponse', (_message.Message,), {
  'DESCRIPTOR' : _REGISTEROBJECTRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.RegisterObjectResponse)
  })
_sym_db.RegisterMessage(RegisterObjectResponse)

StorageServer = _reflection.GeneratedProtocolMessageType('StorageServer', (_message.Message,), {
  'DESCRIPTOR' : _STORAGESERVER,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.StorageServer)
  })
_sym_db.RegisterMessage(StorageServer)

RegisterStorageServerResponse = _reflection.GeneratedProtocolMessageType('RegisterStorageServerResponse', (_message.Message,), {
  'DESCRIPTOR' : _REGISTERSTORAGESERVERRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.RegisterStorageServerResponse)
  })
_sym_db.RegisterMessage(RegisterStorageServerResponse)

StorageClient = _reflection.GeneratedProtocolMessageType('StorageClient', (_message.Message,), {
  'DESCRIPTOR' : _STORAGECLIENT,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.StorageClient)
  })
_sym_db.RegisterMessage(StorageClient)

RegisterStorageClientResponse = _reflection.GeneratedProtocolMessageType('RegisterStorageClientResponse', (_message.Message,), {
  'DESCRIPTOR' : _REGISTERSTORAGECLIENTRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.RegisterStorageClientResponse)
  })
_sym_db.RegisterMessage(RegisterStorageClientResponse)

RegisterStorageGroupResponse = _reflection.GeneratedProtocolMessageType('RegisterStorageGroupResponse', (_message.Message,), {
  'DESCRIPTOR' : _REGISTERSTORAGEGROUPRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.RegisterStorageGroupResponse)
  })
_sym_db.RegisterMessage(RegisterStorageGroupResponse)

UnregisterStorageGroupResponse = _reflection.GeneratedProtocolMessageType('UnregisterStorageGroupResponse', (_message.Message,), {
  'DESCRIPTOR' : _UNREGISTERSTORAGEGROUPRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.UnregisterStorageGroupResponse)
  })
_sym_db.RegisterMessage(UnregisterStorageGroupResponse)

EtcdServer = _reflection.GeneratedProtocolMessageType('EtcdServer', (_message.Message,), {
  'DESCRIPTOR' : _ETCDSERVER,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.EtcdServer)
  })
_sym_db.RegisterMessage(EtcdServer)

UnregisterStorageServerResponse = _reflection.GeneratedProtocolMessageType('UnregisterStorageServerResponse', (_message.Message,), {
  'DESCRIPTOR' : _UNREGISTERSTORAGESERVERRESPONSE,
  '__module__' : 'node_tracker_pb2'
  # @@protoc_insertion_point(class_scope:di_store.node_tracker.UnregisterStorageServerResponse)
  })
_sym_db.RegisterMessage(UnregisterStorageServerResponse)


DESCRIPTOR._options = None

_NODETRACKER = _descriptor.ServiceDescriptor(
  name='NodeTracker',
  full_name='di_store.node_tracker.NodeTracker',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=1263,
  serialized_end=2311,
  methods=[
  _descriptor.MethodDescriptor(
    name='register_storage_server',
    full_name='di_store.node_tracker.NodeTracker.register_storage_server',
    index=0,
    containing_service=None,
    input_type=_STORAGESERVER,
    output_type=_REGISTERSTORAGESERVERRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='unregister_storage_server',
    full_name='di_store.node_tracker.NodeTracker.unregister_storage_server',
    index=1,
    containing_service=None,
    input_type=_STORAGESERVER,
    output_type=_UNREGISTERSTORAGESERVERRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='register_storage_group',
    full_name='di_store.node_tracker.NodeTracker.register_storage_group',
    index=2,
    containing_service=None,
    input_type=_STORAGESERVER,
    output_type=_REGISTERSTORAGEGROUPRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='unregister_storage_group',
    full_name='di_store.node_tracker.NodeTracker.unregister_storage_group',
    index=3,
    containing_service=None,
    input_type=_STORAGESERVER,
    output_type=_UNREGISTERSTORAGEGROUPRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='register_storage_client',
    full_name='di_store.node_tracker.NodeTracker.register_storage_client',
    index=4,
    containing_service=None,
    input_type=_STORAGECLIENT,
    output_type=_REGISTERSTORAGECLIENTRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='register_object',
    full_name='di_store.node_tracker.NodeTracker.register_object',
    index=5,
    containing_service=None,
    input_type=_REGISTEROBJECTREQUEST,
    output_type=_REGISTEROBJECTRESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='server_info',
    full_name='di_store.node_tracker.NodeTracker.server_info',
    index=6,
    containing_service=None,
    input_type=_SERVERINFOREQUEST,
    output_type=_SERVERINFORESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='object_info',
    full_name='di_store.node_tracker.NodeTracker.object_info',
    index=7,
    containing_service=None,
    input_type=_OBJECTINFOREQUEST,
    output_type=_OBJECTINFORESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='object_delete',
    full_name='di_store.node_tracker.NodeTracker.object_delete',
    index=8,
    containing_service=None,
    input_type=_OBJECTDELETEREQUEST,
    output_type=_OBJECTDELETERESPONSE,
    serialized_options=None,
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_NODETRACKER)

DESCRIPTOR.services_by_name['NodeTracker'] = _NODETRACKER

# @@protoc_insertion_point(module_scope)
