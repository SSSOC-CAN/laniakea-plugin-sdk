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

Now that we've satisfied the `Datasource` interface, we need to write our `main()` function.
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

The final step is to build an executable, place it in the `PluginDir` defined in the Laniakea config file and modify our config file to include this plugin.
```yaml
PluginDir: "/home/vagrant/.laniakea"
Plugins:
  - name: "test-plugin"
    type: "datasource"
    exec: "test_plugin" 
    timeout: 30
    maxTimeouts: 3
```
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
The final step is to build an executable, place it in the `PluginDir` defined in the Laniakea config file and modify our config file.
```yaml
PluginDir: "/home/vagrant/.laniakea"
Plugins:
  - name: "test-ctrl-plugin"
    type: "controller"
    exec: "test_ctrl_plugin" 
    timeout: 30
    maxTimeouts: 3
```
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
$ lanicli plugin-command test-ctrl-plugin --command='{ "command": "start-record" }' --frametype="application/json"
{
    "source":  "test-plugin",
    "type":  "application/json",
    "timestamp":  "1663728274439",
    "payload":  "eyJkYXRhIjpbeyJuYW1lIjoiVGVtcGVyYXR1cmU6IDEiLCJ2YWx1ZSI6MjUuMDk5MTcxMzE4MjIzMzAyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMiIsInZhbHVlIjoyNS4wNjcxMjgzNTQ4MDIxNDR9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAzIiwidmFsdWUiOjI1LjE4Nzk5MzYyNTY2Nzh9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA0IiwidmFsdWUiOjI1LjQ2NDEyOTcyNjgxMDA1N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDUiLCJ2YWx1ZSI6MjUuMTU4NDcwMzY3NjI4NjN9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2IiwidmFsdWUiOjI1LjI4MjM2NTU0NjU3Mzk3OH0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDciLCJ2YWx1ZSI6MjUuMDk3NjYzMTM5NTgzMDQ4fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogOCIsInZhbHVlIjoyNS45MjYwODIxOTEyNjk0Nn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDkiLCJ2YWx1ZSI6MjUuOTQ2MDQ1NjgzNjEyMjN9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAxMCIsInZhbHVlIjoyNS42MTA1MjAxNjg3OTg3MTV9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAxMSIsInZhbHVlIjoyNS4xNDkwMjQyMzQ4NDQxODd9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAxMiIsInZhbHVlIjoyNS4wMjMwNTU1NTU1ODQ5Njh9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAxMyIsInZhbHVlIjoyNS43MzkyOTEzNDg4NzMwNTN9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAxNCIsInZhbHVlIjoyNS4wNzMyMzE1MjE3NTI3Nn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDE1IiwidmFsdWUiOjI1Ljg5OTkwNDgwNjI3MjU3M30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDE2IiwidmFsdWUiOjI1LjUzMjAyODIwODczNDk1N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDE3IiwidmFsdWUiOjI1LjgwNzc0NTg0MzA0NTcyNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDE4IiwidmFsdWUiOjI1LjcyMDg1NzY3NjM5MzE4NH0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDE5IiwidmFsdWUiOjI1LjU0NTg0Nzc2MTM4NTM1M30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDIwIiwidmFsdWUiOjI1LjI1MDY2NjgyODQ3MTYzN30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDIxIiwidmFsdWUiOjI1LjM1Nzk0NTI2NDMxNzU3fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMjIiLCJ2YWx1ZSI6MjUuODkxMTg3NDkxMzMzMTM4fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMjMiLCJ2YWx1ZSI6MjUuNzk5NDAzNDg4ODM2NDU0fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMjQiLCJ2YWx1ZSI6MjUuMjY2NTQ4ODQ2NTI0ODAyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMjUiLCJ2YWx1ZSI6MjUuNDQ3MjkxMTg0MDQ4Nzh9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAyNiIsInZhbHVlIjoyNS4xMjgzMzU2OTQyOTM2NTd9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAyNyIsInZhbHVlIjoyNS43NzUwNzg5MTczMjY1NTd9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAyOCIsInZhbHVlIjoyNS4zMTg2NTkwMzIyMTE2Mn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDI5IiwidmFsdWUiOjI1LjQwMzI4OTE2MTM3NTg3N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDMwIiwidmFsdWUiOjI1LjU2NTkyOTI4NjEyMzU0NX0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDMxIiwidmFsdWUiOjI1LjA3MTkwMTI3NjMzMDgwNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDMyIiwidmFsdWUiOjI1LjI4OTg3NzEyNDY0NTU0Nn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDMzIiwidmFsdWUiOjI1LjYwNjcyMjc1NjQ3NzUwNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDM0IiwidmFsdWUiOjI1LjgyNjI2MTE3NTQ1MDc3fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogMzUiLCJ2YWx1ZSI6MjUuNjI5NTg3NzE2MzQ5Mzl9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAzNiIsInZhbHVlIjoyNS4xNjEyMTQ5NDI5NjUxODR9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiAzNyIsInZhbHVlIjoyNS4yMTAwMzA1ODY5NjIyN30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDM4IiwidmFsdWUiOjI1LjcyNjEzMDUwMjQ1ODAzNH0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDM5IiwidmFsdWUiOjI1LjU4NDEwMzc4NzQ1NzIzfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNDAiLCJ2YWx1ZSI6MjUuOTI0MDIwOTg4MzYwNDc1fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNDEiLCJ2YWx1ZSI6MjUuNDY3NTk2MDMxODE5Mzg3fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNDIiLCJ2YWx1ZSI6MjUuNzQxMjk4NTI0MDM3NzUyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNDMiLCJ2YWx1ZSI6MjUuNTAyNTk2ODI1NjExNzY4fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNDQiLCJ2YWx1ZSI6MjUuMjgzNTc4MTQ5MTAyNDV9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA0NSIsInZhbHVlIjoyNS45NDUwMjE5ODAzMzc5OX0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDQ2IiwidmFsdWUiOjI1LjMwODcwMDI3MzE5NjExNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDQ3IiwidmFsdWUiOjI1LjczNTY2ODM3Njg4OTc3M30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDQ4IiwidmFsdWUiOjI1LjM3MDExMzgzODc2ODIzN30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDQ5IiwidmFsdWUiOjI1LjEyNTQ2NjU0NjIzMzk2Mn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDUwIiwidmFsdWUiOjI1LjAwNzY3MTU3NzUwMjA1fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTEiLCJ2YWx1ZSI6MjUuMDEzOTM0NDM1NjgxMzQ2fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTIiLCJ2YWx1ZSI6MjUuNDcwMzQwODE0NTI0ODF9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA1MyIsInZhbHVlIjoyNS41Mzk0ODIyMTkzMDM5OX0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDU0IiwidmFsdWUiOjI1Ljc1MDM4NzM4MjU3ODcyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTUiLCJ2YWx1ZSI6MjUuNDc2OTY2MzA3NzUyODY0fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTYiLCJ2YWx1ZSI6MjUuODYwNzgzODUwNTE4MzI0fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTciLCJ2YWx1ZSI6MjUuMTU2MDQxNTE1MDE0Njc1fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTgiLCJ2YWx1ZSI6MjUuNTQxNjM4NzEwNDA4NDc4fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNTkiLCJ2YWx1ZSI6MjUuMjY4NTY1NjUxMDkyNzR9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2MCIsInZhbHVlIjoyNS45NDc3MjgwMTA2NzQwMDJ9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2MSIsInZhbHVlIjoyNS43Nzc4MzAyNzcwNjYxNzJ9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2MiIsInZhbHVlIjoyNS43Mzc0MTczOTkwNTE1Mn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDYzIiwidmFsdWUiOjI1LjU5NzI3Njc5NzM3MzExfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNjQiLCJ2YWx1ZSI6MjUuMDczMjkzOTk4NjkzMTJ9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2NSIsInZhbHVlIjoyNS41MzkxNzg2NzA4MzQyODJ9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA2NiIsInZhbHVlIjoyNS42ODYzOTYwMDI2NzI5N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDY3IiwidmFsdWUiOjI1LjI2ODQzMTE3NjMwOTgwN30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDY4IiwidmFsdWUiOjI1LjA0MTMyMDUxOTc4ODIyMn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDY5IiwidmFsdWUiOjI1LjY0ODA0NTEyNjY0NTE3Mn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDcwIiwidmFsdWUiOjI1LjQxMzA0ODIxMDgzNDIyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogNzEiLCJ2YWx1ZSI6MjUuOTA3NDIzMTgxNTI3NDl9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA3MiIsInZhbHVlIjoyNS4wNzMxMzk4MDU3Mjc0MDh9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA3MyIsInZhbHVlIjoyNS43MDI0NTk5NzYyMzA4NzN9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA3NCIsInZhbHVlIjoyNS4xNzEzMDExMzM1ODY2M30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDc1IiwidmFsdWUiOjI1Ljg4ODg2NDIzNTUyMDkxNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDc2IiwidmFsdWUiOjI1LjkwNjM2OTUyNTYzMjg1OH0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDc3IiwidmFsdWUiOjI1LjYyODkzNzQ4ODE3MTU0OH0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDc4IiwidmFsdWUiOjI1LjYzMDk3MjgxMzcxODI3M30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDc5IiwidmFsdWUiOjI1LjAxNTk1NzM1NDg5NzQ4fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogODAiLCJ2YWx1ZSI6MjUuMjIyOTc5NTgxMzQwMTAyfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogODEiLCJ2YWx1ZSI6MjUuNjQ0NjM3MTc5NTM2MTM3fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogODIiLCJ2YWx1ZSI6MjUuNTE3NjgyNDQzODYzMDQ1fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogODMiLCJ2YWx1ZSI6MjUuMDg2MzYzMzE5NTQ0OTg2fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogODQiLCJ2YWx1ZSI6MjUuMjQxNzY4NDQ0NTEyNTh9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA4NSIsInZhbHVlIjoyNS43NzQyMDM5Mzc3OTc2N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDg2IiwidmFsdWUiOjI1LjcyODQ5NjE0MzY0NjIzNn0seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDg3IiwidmFsdWUiOjI1LjU5NzAwNjIxNzU3MDAzM30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDg4IiwidmFsdWUiOjI1LjI1MzAxODgzNzcxODI4N30seyJuYW1lIjoiVGVtcGVyYXR1cmU6IDg5IiwidmFsdWUiOjI1Ljg2NDg0NDQ1MTMyMDExfSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogOTAiLCJ2YWx1ZSI6MjUuNTA2NzY0MDc5NDYxMjA2fSx7Im5hbWUiOiJUZW1wZXJhdHVyZTogOTEiLCJ2YWx1ZSI6MjUuOTEzNzU2Nzk2MTEzNDV9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA5MiIsInZhbHVlIjoyNS44NTQ2NTEwNzI3NTI2OTV9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA5MyIsInZhbHVlIjoyNS42ODgzODA0MDY1NjM3NzN9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA5NCIsInZhbHVlIjoyNS4zMDQ5ODQ3MDI1OTExOTV9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA5NSIsInZhbHVlIjoyNS4yNzM3MzM3NjQ5NTMzMTd9LHsibmFtZSI6IlRlbXBlcmF0dXJlOiA5NiIsInZhbHVlIjoyNS43NTA5NTA5Njc0NDgzNTd9XX0="
}
```