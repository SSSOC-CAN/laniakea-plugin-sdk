# Python Plugin

This guide gives step by step instructions on how to make basic `Datasource` and `Controller` plugins in Python for Laniakea. Laniakea is a tool to easily connect hardware sensors and controllers uner one common interface.

## Step 1: Plugin file and Dependencies

In order to write and compile a plugin in Python, we must have Python 3.7 or higher installed. Python virtual environment is also recommended. Let's create a virtual environment in a new folder and create a new file called `plugin.py`.
```
$ python -m venv venv
$ source venv/bin/activate
(venv) $ touch plugin.py
(venv) $ pip install grpcio grpcio-tools grpcio-health-checking laniakea-plugin-sdk pyinstaler
```

## Step 2: Datsource Plugin Implementation

### Versioning

When Laniakea initializes the plugins, it starts by first pushing its version string to all plugins using the `PushVersion` function defined in the `Datasource` interface. It then retrives the plugin version string using `GetVersion` again defined in the `Datasource` interface. Thankfully, the implementations of these functions have already been written and are given in the SDK. The only thing we must do is define our Laniakea version constraint and our plugin version.

```python
pluginVersion = "1.0.0"
laniVersionConstraint = "0.2.0"
```

This means that if the plugin is being run by a version of Laniakea other than `0.2.0`, it will raise an error.

### Datasource Interface Implementation

Laniakea datasource plugins need to satisfy the `Datasource` interface. 
```go
type Datasource interface{
    StartRecord() (chan *proto.Frame, error)
    StopRecord() error
    Stop() error
    PushVersion(versionNumber string) error
    GetVersion() (string, error)
}
```
`PushVersion` and `GetVersion` are already implemented but we still need to implement `StartRecord`, `StopRecord` and `Stop`. Our datasource plugin is going to be a random number generator. We want `StopRecord` and `Stop` to be able to detect if we are currently producing random numbers and give them the ability to stop that process.

```python
from laniakea_plugin_sdk.datasource import DatasourceBase
from laniakea_plugin_sdk.datasource import Serve
import random
import string
import time

import laniakea_plugin_sdk.proto.plugin_pb2 as plugin_pb2
import laniakea_plugin_sdk.proto.plugin_pb2_grpc as plugin_pb2_grpc

class Datasource(DatasourceBase):
    def __init__(self, version, laniVersionConstraint):
        super(Datasource, self).__init__(version=version, laniVersionConstraint=laniVersionConstraint) #lets make sure we initialize the parent class here
        self.quit = False

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
            time.sleep(1)
    
    def StopRecord(self, request, context):
        self.quit = True
        return plugin_pb2.Empty()

    def Stop(self, request, context):
        self.quit = True
        return plugin_pb2.Empty()
```

### Final steps

Now that we've satisfied the `Datasource` interface, we need to write our `if __name__ == '__main__':` block.
```python
if __name__ == '__main__':
    Serve(Datasource(version=pluginVersion, laniVersionConstraint=laniVersionConstraint))
```

Finally, we need to build the executable. To build executables in Python, we are using `pyinstaller`. Be sure to place the executable in the `PluginDir` or change the location of `PluginDir` to be the parent directory of our newly created executable file. We must also modify the Laniakea config file to include our new plugin.
```yaml
PluginDir: "/home/vagrant/.laniakea"
Plugins:
  - name: "test-plugin"
    type: "datasource"
    exec: "plugin" 
    timeout: 30
    maxTimeouts: 3
```
```
(venv) $ pyinstaller -F plugin.py
(venv) $ mv dist/plugin ~/.laniakea
(venv) $ laniakea

4:41PM [INFO]    Starting Daemon...
4:41PM [INFO]    Daemon succesfully started. Version: 0.2.0
4:41PM [INFO]    Loading TLS configuration...
4:41PM [INFO]    TLS configuration successfully loaded.
4:41PM [INFO]    RPC Server Initialized.
4:41PM [INFO]    Initializing plugins...
4:41PM [DEBUG]   starting plugin: path=/home/vagrant/.laniakea/plugin args=[/home/vagrant/.laniakea/plugin]  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin started: path=/home/vagrant/.laniakea/plugin pid=7568  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   waiting for RPC address: path=/home/vagrant/.laniakea/plugin  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   using plugin: version=1  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin address: address=127.0.0.1:10000 network=tcp timestamp=2022-09-21T16:41:41.964-0400  plugin=test-plugin subsystem=PLGN
4:41PM [TRACE]   waiting for stdio data plugin=test-plugin subsystem=PLGN
4:41PM [INFO]    registered plugin test-plugin version: 1.0.0 subsystem=PLGN
4:41PM [INFO]    Plugins initialized
4:41PM [INFO]    RPC subservices instantiated and registered successfully.
4:41PM [INFO]    Opening database...
4:41PM [INFO]    Database successfully opened.
4:41PM [INFO]    Initializing unlocker service...
4:41PM [INFO]    Unlocker service initialized.
4:41PM [INFO]    Starting RPC server... subsystem=RPCS
4:41PM [INFO]    RPC server started subsystem=RPCS
4:41PM [INFO]    gRPC listening on [::]:7777 subsystem=RPCS
4:41PM [INFO]    REST proxy started and listening at 0.0.0.0:8080 subsystem=RPCS
4:41PM [INFO]    Waiting for password. Use `lanicli setpassword` to set a password for the first time, `lanicli login` to unlock the daemon with an existing password, or `lanicli changepassword` to change the existing password and unlock the daemon.
```
Then from `lanicli`
```
$ lanicli login
$ lanicli plugin-startrecord test-plugin
{
}
```

## Step 3: Controller Plugin Implementation

### Versioning and Controller Interface Implementation

Like the `Datasource` plugin, we must satisfy the `Controller` interface.

```go
type Controller interface{
    Command(*proto.Frame) (chan *proto.Frame, error)
    Stop() error
    PushVersion(versionNumber string) error
    GetVersion() (string, error)
}
```
And like the `Datasource`, the SDK already provides a partially satisfied interface in the form of `ControllerBase` struct with the `PushVersion` and `GetVersion` functions implemented. All that we must do, is specify our Laniakea version constraint, our plugin version and implement `Command()` and `Stop()`.

```python
import json
from laniakea_plugin_sdk.controller import ControllerBase
from laniakea_plugin_sdk.controller import Serve
import random
import string
import time

import laniakea_plugin_sdk.proto.plugin_pb2 as plugin_pb2
import laniakea_plugin_sdk.proto.plugin_pb2_grpc as plugin_pb2_grpc

pluginName = "test-ctrl-pyplugin"
pluginVersion = "1.0.0"
laniVersionConstraint = "0.2.0"

class Controller(ControllerBase):
    def __init__(self, version, laniVersionConstraint):
        super(Controller, self).__init__(version=version, laniVersionConstraint=laniVersionConstraint)
        self.quit = False

    def GetVersion(self, request, context):
        return super(Datasource, self).GetVersion(request, context)

    def PushVersion(self, request, context):
        return super(Datasource, self).PushVersion(request, context)

    def Command(self, request, context):
        if request.type == "application/json":
            cmd = json.loads(request.payload.decode('utf-8'))
            if cmd['command'] == "start-record":
                while True:
                    if self.quit:
                        break
                    yield plugin_pb2.Frame(
                        source=pluginName,
                        type="application/json",
                        timestamp=int(round(time.time() * 1000)),
                        payload= bytes(''.join(random.choice(string.ascii_lowercase) for i in range(10)), 'utf-8'),
                    )
                    time.sleep(1)

    def Stop(self, request, context):
        self.quit = True
        return plugin_pb2.Empty()


if __name__ == '__main__':
    Serve(Controller(version=pluginVersion, laniVersionConstraint=laniVersionConstraint))
```

The final step is to build an executable, place it in the `PluginDir` defined in the Laniakea config file and modify the config file.
```yaml
PluginDir: "/home/vagrant/.laniakea"
Plugins:
  - name: "test-ctrl-plugin"
    type: "controller"
    exec: "ctrl_plugin" 
    timeout: 30
    maxTimeouts: 3
```
```
(venv) $ pyinstaller -F ctrl_plugin.py
(venv) $ mv dist/ctrl_plugin.py ~/.laniakea
(venv) $ laniakea
4:41PM [INFO]    Starting Daemon...
4:41PM [INFO]    Daemon succesfully started. Version: 0.2.0
4:41PM [INFO]    Loading TLS configuration...
4:41PM [INFO]    TLS configuration successfully loaded.
4:41PM [INFO]    RPC Server Initialized.
4:41PM [INFO]    Initializing plugins...
4:41PM [DEBUG]   starting plugin: path=/home/vagrant/.laniakea/ctrl_plugin args=[/home/vagrant/.laniakea/ctrl_plugin]  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin started: path=/home/vagrant/.laniakea/ctrl_plugin pid=7568  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   waiting for RPC address: path=/home/vagrant/.laniakea/ctrl_plugin  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   using plugin: version=1  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin address: address=127.0.0.1:10000 network=tcp timestamp=2022-09-21T16:41:41.964-0400  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [TRACE]   waiting for stdio data plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [INFO]    registered plugin test-ctrl-plugin version: 1.0.0 subsystem=PLGN
4:41PM [INFO]    Plugins initialized
4:41PM [INFO]    RPC subservices instantiated and registered successfully.
4:41PM [INFO]    Opening database...
4:41PM [INFO]    Database successfully opened.
4:41PM [INFO]    Initializing unlocker service...
4:41PM [INFO]    Unlocker service initialized.
4:41PM [INFO]    Starting RPC server... subsystem=RPCS
4:41PM [INFO]    RPC server started subsystem=RPCS
4:41PM [INFO]    gRPC listening on [::]:7777 subsystem=RPCS
4:41PM [INFO]    REST proxy started and listening at 0.0.0.0:8080 subsystem=RPCS
4:41PM [INFO]    Waiting for password. Use `lanicli setpassword` to set a password for the first time, `lanicli login` to unlock the daemon with an existing password, or `lanicli changepassword` to change the existing password and unlock the daemon.
```
Then from `lanicli`
```
$ lanicli login
$ lanicli plugin-command test-ctrl-plugin --command='{ "command": "start-record" }' --frametype="application/json"
{
    "source":  "test-ctrl-pyplugin",
    "type":  "application/json",
    "timestamp":  "1663728274439",
    "payload":  "eyJkYXRhIj..."
}
```