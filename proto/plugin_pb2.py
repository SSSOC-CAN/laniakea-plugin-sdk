# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: plugin.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0cplugin.proto\x12\x05proto\"\x07\n\x05\x45mpty\"I\n\x05\x46rame\x12\x0e\n\x06source\x18\x01 \x01(\t\x12\x0c\n\x04type\x18\x02 \x01(\t\x12\x11\n\ttimestamp\x18\x03 \x01(\x03\x12\x0f\n\x07payload\x18\x04 \x01(\x0c\" \n\rVersionNumber\x12\x0f\n\x07version\x18\x01 \x01(\t2\xec\x01\n\nDatasource\x12+\n\x0bStartRecord\x12\x0c.proto.Empty\x1a\x0c.proto.Frame0\x01\x12(\n\nStopRecord\x12\x0c.proto.Empty\x1a\x0c.proto.Empty\x12\"\n\x04Stop\x12\x0c.proto.Empty\x1a\x0c.proto.Empty\x12\x31\n\x0bPushVersion\x12\x14.proto.VersionNumber\x1a\x0c.proto.Empty\x12\x30\n\nGetVersion\x12\x0c.proto.Empty\x1a\x14.proto.VersionNumber2\xbe\x01\n\nController\x12\"\n\x04Stop\x12\x0c.proto.Empty\x1a\x0c.proto.Empty\x12\'\n\x07\x43ommand\x12\x0c.proto.Frame\x1a\x0c.proto.Frame0\x01\x12\x31\n\x0bPushVersion\x12\x14.proto.VersionNumber\x1a\x0c.proto.Empty\x12\x30\n\nGetVersion\x12\x0c.proto.Empty\x1a\x14.proto.VersionNumberB0Z.github.com/SSSOC-CAN/laniakea-plugin-sdk/protob\x06proto3')



_EMPTY = DESCRIPTOR.message_types_by_name['Empty']
_FRAME = DESCRIPTOR.message_types_by_name['Frame']
_VERSIONNUMBER = DESCRIPTOR.message_types_by_name['VersionNumber']
Empty = _reflection.GeneratedProtocolMessageType('Empty', (_message.Message,), {
  'DESCRIPTOR' : _EMPTY,
  '__module__' : 'plugin_pb2'
  # @@protoc_insertion_point(class_scope:proto.Empty)
  })
_sym_db.RegisterMessage(Empty)

Frame = _reflection.GeneratedProtocolMessageType('Frame', (_message.Message,), {
  'DESCRIPTOR' : _FRAME,
  '__module__' : 'plugin_pb2'
  # @@protoc_insertion_point(class_scope:proto.Frame)
  })
_sym_db.RegisterMessage(Frame)

VersionNumber = _reflection.GeneratedProtocolMessageType('VersionNumber', (_message.Message,), {
  'DESCRIPTOR' : _VERSIONNUMBER,
  '__module__' : 'plugin_pb2'
  # @@protoc_insertion_point(class_scope:proto.VersionNumber)
  })
_sym_db.RegisterMessage(VersionNumber)

_DATASOURCE = DESCRIPTOR.services_by_name['Datasource']
_CONTROLLER = DESCRIPTOR.services_by_name['Controller']
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z.github.com/SSSOC-CAN/laniakea-plugin-sdk/proto'
  _EMPTY._serialized_start=23
  _EMPTY._serialized_end=30
  _FRAME._serialized_start=32
  _FRAME._serialized_end=105
  _VERSIONNUMBER._serialized_start=107
  _VERSIONNUMBER._serialized_end=139
  _DATASOURCE._serialized_start=142
  _DATASOURCE._serialized_end=378
  _CONTROLLER._serialized_start=381
  _CONTROLLER._serialized_end=571
# @@protoc_insertion_point(module_scope)