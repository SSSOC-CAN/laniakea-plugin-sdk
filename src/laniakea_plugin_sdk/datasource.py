"""
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/09/07
"""

import laniakea_plugin_sdk.proto.plugin_pb2 as plugin_pb2
import laniakea_plugin_sdk.proto.plugin_pb2_grpc as plugin_pb2_grpc

from concurrent import futures
import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer
from packaging import version
import sys
import time

defaultPort = 1234

class DatasourceBase(plugin_pb2_grpc.DatasourceServicer):
    """Base Implementation of Datasource Plugin"""
    def __init__(self, version, laniVersionConstraint):
        super(DatasourceBase, self).__init__()
        self.setPluginVersion(version)
        self.setVersionConstraints(laniVersionConstraint)

    def setVersionConstraints(self, verStr: str):
        if isinstance(version.parse(verStr), version.LegacyVersion):
            raise ValueError("invalid version string")
        self.laniVersionConstraint = verStr
    
    def setPluginVersion(self, verStr: str):
        if isinstance(version.parse(verStr), version.LegacyVersion):
            raise ValueError("invalid version string")
        self.version = verStr
    
    def GetVersion(self, request, context):
        if not self.version:
            raise ValueError("plugin version not set")
        return plugin_pb2.VersionNumber(version=self.version)

    def PushVersion(self, request, context):
        laniV = version.parse(request.version)
        if isinstance(laniV, version.LegacyVersion):
            raise ValueError("invalid version string")
        plugVConstraint = version.parse(self.laniVersionConstraint)
        if laniV < plugVConstraint:
            raise ValueError("plugin requires a different version of laniakea")
        self.laniVersion = request.version

def Serve(servicer):
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    plugin_pb2_grpc.add_DatasourceServicer_to_server(servicer, server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    port = defaultPort
    while True:
        srv_str = f'127.0.0.1:{port}'
        try:
            server.add_insecure_port(srv_str)
        except:
            port += 1
        else:
            break
    server.start()

    # Output information
    print(f'1|1|tcp|{srv_str}|grpc')
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)