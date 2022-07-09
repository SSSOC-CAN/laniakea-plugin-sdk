/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/07/07
*/

package laniakea_sdk

import (
	"context"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var (
	HandshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "LANIAKEA_PLUGIN_MAGIC_COOKIE",
		MagicCookieValue: "a56e5daaa516e17d3d4b3d4685df9f8ca59c62c2d818cd5a7df13c039f134e16",
	}
)

// Datasource interface describes an interface for plugins which will only produce streams of data
type Datasource interface {
	StartRecord() (chan *proto.Frame, error)
	StopRecord() error
	Stop() error
	PushVersion(versionNumber string) error // This method pushes the version of Laniakea to the plugin. Plugin can then specify a minimum version of laniakea to run properly
	GetVersion() (string, error)            // This method gets the version number from the plugin. Needed if plugins rely on other plugins and specific versions are needed
}

// Controller interface describes an interface for plugins which produce data but also act as controllers
type Controller interface {
	Stop() error
	Command(*proto.Frame) (chan *proto.Frame, error)
	PushVersion(versionNumber string) error // This method pushes the version of Laniakea to the plugin. Plugin can then specify a minimum version of laniakea to run properly
	GetVersion() (string, error)            // This method gets the version number from the plugin. Needed if plugins rely on other plugins and specific versions are needed
}

type DatasourcePlugin struct {
	plugin.Plugin
	Impl Datasource
}

type ControllerPlugin struct {
	plugin.Plugin
	Impl Controller
}

// GRPCServer implements the plugin.Plugin interface in the go-plugin package
func (p *DatasourcePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterDatasourceServer(s, &DatasourceGRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient implements the plugin.Plugin interface in the go-plugin package
func (p *DatasourcePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &DatasourceGRPCClient{client: proto.NewDatasourceClient(c)}, nil
}

// GRPCServer implements the plugin.Plugin interface in the go-plugin package
func (p *ControllerPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterControllerServer(s, &ControllerGRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient implements the plugin.Plugin interface in the go-plugin package
func (p *ControllerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &ControllerGRPCClient{client: proto.NewControllerClient(c)}, nil
}
