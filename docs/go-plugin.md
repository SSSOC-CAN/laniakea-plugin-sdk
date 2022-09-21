# Go Plugin

This guide gives step by step instructions on how to make both basic `Datasource` and `Controller` plugins in Go for Laniakea. Laniakea is a tool to easily connect hardware sensors and controllers under one common interface.

## Step 1: Plugin file and Dependencies

In order to write and compile a plugin in Go, we must have Go installed. Version 1.17 or higher is required.
Let us create a new folder called `test_plugin` and initialize a new go project. 
```
$ mkdir test_plugin && cd test_plugin
$ touch main.go
$ go mod init 
$ go get github.com/SSSOC-CAN/laniakea-plugin/sdk
$ go get github.com/hashicorp/go-plugin
```

## Step 2: Datasource Plugin Implementation

### Versioning

When Laniakea initializes the plugins, it starts by first pushing its version string to all plugins using the `PushVersion` function defined in the `Datasource` interface. It then retrives the plugin version string using `GetVersion` again defined in the `Datasource` interface. Thankfully, the implementations of these functions have already been written and are given in the SDK. The only thing we must do is define our Laniakea version constraint and our plugin version.

```go
package main

var (
    pluginVersion = "1.0.0"
    laniVersionConstraint = ">= 0.2.0"
)
```
This means that if the plugin is being run by a version of Laniakea older than `0.2.0`, it will raise an error.

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
`PushVersion` and `GetVersion` are already implemented but we still need to implement `StartRecord`, `StopRecord` and `Stop`. Our datasource plugin is going to be a random number generator which runs in a go routine. We want `StopRecord` and `Stop` to be able to detect if we are currently producing random numbers and give them the ability to kill that go routine.

```go
import (
    "fmt"
    "math/rand"
    "sync"
    "sync/atomic"
    "time"

    sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
    "github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
)

type DatasourceExample struct {
    sdk.DatasourceBase
    Recording int32 // used atomically
    quitChan chan struct{}
    sync.WaitGroup
}

func (d *DatasourceExample) StartRecord() (chan *proto.Frame, error) {
    if ok := atomic.CompareAndSwapInt32(&d.Recording, 0, 1); !ok {
        return nil, fmt.Errorf("could not start recording: recording already started")
    }
    frameChan := make(chan *proto.Frame)
	e.Add(1)
	go func() {
		defer e.Done()
		defer close(frameChan)
		time.Sleep(1 * time.Second) // sleep for a second while laniakea sets up the plugin
		for {
			select {
			case <-e.quitChan:
				return
			default:
				rnd := make([]byte, 8)
				_, err := rand.Read(rnd)
				if err != nil {
					log.Println(err)
				}
				frameChan <- &proto.Frame{
					Source:    "test-plugin", // this can be anything you want. Ideally this should be the name of your plugin or some other useful information for whoever receives this data
					Type:      "application/json", // again this can be anything. If you have multiple plugins running using different types of data, this is how you identify it's format.
					Timestamp: time.Now().UnixMilli(),
					Payload:   rnd, // The actual data, this will be a byte slice of your data. JSON format is advised here
				}
				time.Sleep(1 * time.Second) // This isn't necessary, I just add a 1 second delay for convenience
			}
		}
	}()
	return frameChan, nil
}

func (d *DatasourceExample) StopRecord() error {
    if ok := atomic.CompareAndSwapInt32(&d.Recording, 1, 0); !ok {
        return nil, fmt.Errorf("could not stop recording: recording already stopped")
    }
    e.quitChan <- struct{}{}
    return nil
}

func (d *DatasourceExample) Stop() error {
    close(e.quitChan) // here we don't care if we close the channel since Laniakea will also kill the plugin after this executes or the plugin times out
    e.Wait() // no goroutine leaks this way
    return nil
}
```

### Final steps

Now that we've satisfied the `Datasource` inteface, we need to write our `main()` function.
```go
import (
    ...
    "github.com/hashicorp/go-plugin"
)

func main() {
    impl := &DatasourceExample{quitChan: make(chan struct{})}
    impl.SetPluginVersion(pluginVersion) // set the plugin version before GetVersion is called
    impl.SetVersionConstraints(laniVersionConstraint) // set laniakea version constraint before PushVersion is called
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "test-plugin": &sdk.DatasourcePlugin{Impl: impl}, 
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```
For more information about the underlying pluging package we use `hashicorp/go-plugin` please visit their [repository](https://github.com/hashicorp/go-plugin).

The final step is to build an executable and place it in the `PluginDir` defined in the `Laniakea` config file.
```
$ go build -o test_plugin main.go
$ mv test_plugin ~/.laniakea          # or wherever your PluginDir is
$ laniakea
4:41PM [INFO]    Starting Daemon...
4:41PM [INFO]    Daemon succesfully started. Version: 0.2.0
4:41PM [INFO]    Loading TLS configuration...
4:41PM [INFO]    TLS configuration successfully loaded.
4:41PM [INFO]    RPC Server Initialized.
4:41PM [INFO]    Initializing plugins...
4:41PM [DEBUG]   starting plugin: path=/home/vagrant/.laniakea/test_plugin args=[/home/vagrant/.laniakea/test_plugin]  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin started: path=/home/vagrant/.laniakea/test_plugin pid=7568  plugin=test-plugin subsystem=PLGN
4:41PM [DEBUG]   waiting for RPC address: path=/home/vagrant/.laniakea/test_plugin  plugin=test-plugin subsystem=PLGN
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

```go
package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "math/rand"
	"strconv"
    "sync"
    "sync/atomic"
    "time"

    sdk "github.com/SSSOC-CAN/laniakea-plugin-sdk"
	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
)

var (
	pluginVersion         = "1.0.0"
	laniVersionConstraint = ">= 0.2.0"
)

type ControllerExample struct {
	sdk.ControllerBase
	quitChan     chan struct{}
	tempSetPoint float64
	avgTemp      float64
	recording    int32 // used atomically
	sync.WaitGroup
	sync.RWMutex
}

// We define DemoPayload and DemoFrame structs to define the format of our outbound data
type DemoPayload struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type DemoFrame struct {
	Data []DemoPayload `json:"data"`
}

// We define this struct because we want the req that's passed from Laniakea to the plugin to have a JSON format
type CtrlCommand struct {
	Command string            `json:"command"`
	Args    map[string]string `json:"args"`
}

func (e *ControllerExample) Command(req *proto.Frame) (chan *proto.Frame, error) {
	if req == nil {
		return nil, errors.New("request arg is nil")
	}
	frameChan := make(chan *proto.Frame)
	switch req.Type {
	case "application/json":
		var cmd CtrlCommand
		err := json.Unmarshal(req.Payload, &cmd)
		if err != nil {
			return nil, err
		}
		switch cmd.Command {
		case "set-temperature":
			if atomic.LoadInt32(&e.recording) == 1 {
				if setPtStr, ok := cmd.Args["temp_set_point"]; ok {
					if v, err := strconv.ParseFloat(setPtStr, 64); err == nil {
						e.Add(1)
						go func() {
							defer e.Done()
							defer close(frameChan)
							time.Sleep(1 * time.Second)
							e.RLock()
							frameChan <- &proto.Frame{
								Source:    "test-ctrl-plugin",
								Type:      "application/json",
								Timestamp: time.Now().UnixMilli(),
								Payload:   []byte(fmt.Sprintf(`{ "average_temperature": %f, "current_set_point": %f}`, e.avgTemp, e.tempSetPoint)),
							}
							e.RUnlock()
							e.Lock()
							e.tempSetPoint = v
							e.Unlock()
							e.RLock()
							frameChan <- &proto.Frame{
								Source:    "test-ctrl-plugin",
								Type:      "application/json",
								Timestamp: time.Now().UnixMilli(),
								Payload:   []byte(fmt.Sprintf(`{ "average_temperature": %f, "current_set_point": %f}`, e.avgTemp, e.tempSetPoint)),
							}
							e.RUnlock()
						}()
						return frameChan, nil
					}
					return nil, errors.New("not recording")
				}
			}
			return nil, errors.New("invalid temperature set point type")
		case "start-record":
			if ok := atomic.CompareAndSwapInt32(&e.recording, 0, 1); !ok {
				return nil, errors.New("already recording")
			}
			e.Add(1)
			go func() {
				defer e.Done()
				defer close(frameChan)
				for {
					select {
					case <-e.quitChan:
						return
					default:
						data := []DemoPayload{}
						df := DemoFrame{}
						var cumTemp float64
						for i := 0; i < 96; i++ {
							e.RLock()
							v := (rand.Float64()*1 + e.tempSetPoint)
							e.RUnlock()
							n := fmt.Sprintf("Temperature: %v", i+1)
							data = append(data, DemoPayload{Name: n, Value: v})
							cumTemp += v
						}
						e.Lock()
						e.avgTemp = cumTemp / float64(len(data))
						e.Unlock()
						df.Data = data[:]
						// transform to json string
						b, err := json.Marshal(&df)
						if err != nil {
							log.Println(err)
							return
						}
						frameChan <- &proto.Frame{
							Source:    "test-ctrl-plugin",
							Type:      "application/json",
							Timestamp: time.Now().UnixMilli(),
							Payload:   b,
						}
						time.Sleep(1 * time.Second)
					}
				}
			}()
			return frameChan, nil
        case "stop-record":
            if ok := atomic.CompareAndSwapInt32(&e.recording, 1, 0); !ok {
				return nil, errors.New("already stopped recording")
			}
            e.Add(1)
            go func() {
                defer e.Done()
                defer close(frameChan)
                e.quitChan <- struct{}{}
            }()
            return frameChan, nil
		default:
			return nil, errors.New("invalid command")
		}
	default:
		return nil, errors.New("invalid frame type")
	}
}

func (e *ControllerExample) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&e.recording, 1, 0); !ok {
		return errors.New("already stopped recording")
	}
	close(e.quitChan)
	e.Wait()
	return nil
}
```

### Final steps

Now that we've satisfied the `Controller` inteface, we need to write our `main()` function.
```go
import (
    ...
    "github.com/hashicorp/go-plugin"
)

func main() {
    impl := &ControllerExample{quitChan: make(chan struct{}), tempSetPoint: 25.0}
    impl.SetPluginVersion(pluginVersion) // set the plugin version before GetVersion is called
    impl.SetVersionConstraints(laniVersionConstraint) // set laniakea version constraint before PushVersion is called
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: sdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "test-ctrl-plugin": &sdk.ControllerPlugin{Impl: impl}, 
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```
The final step is to build an executable and place it in the `PluginDir` defined in the `Laniakea` config file.

```
$ go build -o test_ctrl_plugin main.go
$ mv test_ctrl_plugin ~/.laniakea          # or wherever your PluginDir is
$ laniakea
4:41PM [INFO]    Starting Daemon...
4:41PM [INFO]    Daemon succesfully started. Version: 0.2.0
4:41PM [INFO]    Loading TLS configuration...
4:41PM [INFO]    TLS configuration successfully loaded.
4:41PM [INFO]    RPC Server Initialized.
4:41PM [INFO]    Initializing plugins...
4:41PM [DEBUG]   starting plugin: path=/home/vagrant/.laniakea/test_ctrl_plugin args=[/home/vagrant/.laniakea/test_ctrl_plugin]  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   plugin started: path=/home/vagrant/.laniakea/test_ctrl_plugin pid=7568  plugin=test-ctrl-plugin subsystem=PLGN
4:41PM [DEBUG]   waiting for RPC address: path=/home/vagrant/.laniakea/test_ctrl_plugin  plugin=test-ctrl-plugin subsystem=PLGN
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
$ lanicli plugin-command test-plugin --command='{ "command": "start-record" }' --frametype="application/json"
{
    
}
```