from laniakea_plugin_sdk.datasource import DatasourceBase
from laniakea_plugin_sdk.datasource import Serve
import time

import laniakea_plugin_sdk.proto.plugin_pb2 as plugin_pb2
import laniakea_plugin_sdk.proto.plugin_pb2_grpc as plugin_pb2_grpc

pluginName = "test-rng-pyplugin"
pluginVersion = "0.1.0"
laniVersionConstraint = "0.2.0"

class Datasource(DatasourceBase):
    def __init__(self, version, laniVersionConstraint):
        super(Datasource, self).__init__(version=version, laniVersionConstraint=laniVersionConstraint)

    def GetVersion(self, request, context):
        return super(Datasource, self).GetVersion(request, context)

    def PushVersion(self, request, context):
        return super(Datasource, self).PushVersion(request, context)

    def StartRecord(self, request, context):
        while True:
            if self.quit:
                break
            yield plugin_pb2.Frame(
                source=pluginName,
                type="application/json",
                timestamp=int(round(time.time() * 1000)),
                payload= bytes(''.join(random.choice(string.ascii_lowercase) for i in range(10)), 'utf-8'),
            )
    
    def StopRecord(self, request, context):
        self.quit = True

    def Stop(self, request, context):
        self.quit = True


if __name__ == '__main__':
    Serve(Datasource(version=pluginVersion, laniVersionConstraint=laniVersionConstraint))