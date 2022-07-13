/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/07/07
*/

package laniakea_sdk

import (
	"context"
	"errors"
	"io"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/hashicorp/go-version"
)

const (
	ErrPluginVersionNotSet     = bg.Error("plugin version not set")
	ErrLaniakeaVersionMismatch = bg.Error("plugin requires a different version of laniakea")
)

type DatasourceGRPCClient struct{ client proto.DatasourceClient }

type DatasourceGRPCServer struct {
	proto.UnimplementedDatasourceServer
	Impl Datasource
}

// StartRecord implements the Datasource interface method StartRecord
func (c *DatasourceGRPCClient) StartRecord() (chan *proto.Frame, error) {
	stream, err := c.client.StartRecord(context.Background(), &proto.Empty{})
	if err != nil {
		return nil, err
	}
	// sometimes the first stream receive is an error
	frameChan := make(chan *proto.Frame)
	frame, err := stream.Recv()
	if frame == nil || err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	go func() {
		defer close(frameChan)
		for {
			frame, err := stream.Recv()
			if frame == nil || err == io.EOF {
				return
			}
			if err != nil {
				break
			}
			frameChan <- frame
		}
	}()
	return frameChan, nil
}

// StopRecord implements the Datasource interface method StopRecord
func (c *DatasourceGRPCClient) StopRecord() error {
	_, err := c.client.StopRecord(context.Background(), &proto.Empty{})
	if err != nil {
		return err
	}
	return nil
}

// Stop implements the Datasource interface method Stop
func (c *DatasourceGRPCClient) Stop() error {
	_, err := c.client.Stop(context.Background(), &proto.Empty{})
	if err != nil {
		return err
	}
	return nil
}

// PushVersion implements the Datasource interface method PushVersion
func (c *DatasourceGRPCClient) PushVersion(versionNumber string) error {
	_, err := c.client.PushVersion(context.Background(), &proto.VersionNumber{Version: versionNumber})
	if err != nil {
		return err
	}
	return nil
}

// GetVersion implements the Datasource interface method GetVersion
func (c *DatasourceGRPCClient) GetVersion() (string, error) {
	resp, err := c.client.GetVersion(context.Background(), &proto.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Version, nil
}

// StartRecord implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) StartRecord(_ *proto.Empty, stream proto.Datasource_StartRecordServer) error {
	frameChan, err := s.Impl.StartRecord()
	if err != nil {
		return err
	}
	for {
		select {
		case frame := <-frameChan:
			if frame == nil {
				return nil
			}
			if err := stream.Send(frame); err != nil {
				return err
			}
		case <-stream.Context().Done():
			if errors.Is(stream.Context().Err(), context.Canceled) {
				return nil
			}
			return stream.Context().Err()
		}
	}
}

// StopRecord implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) StopRecord(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	err := s.Impl.StopRecord()
	return &proto.Empty{}, err
}

// Stop implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) Stop(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	err := s.Impl.Stop()
	return &proto.Empty{}, err
}

// PushVersion implements the Datasource gRPC server interface
func (s *DatasourceGRPCServer) PushVersion(ctx context.Context, req *proto.VersionNumber) (*proto.Empty, error) {
	err := s.Impl.PushVersion(req.Version)
	return &proto.Empty{}, err
}

// GetVersion implements the Datsource gRPC server interface
func (s *DatasourceGRPCServer) GetVersion(ctx context.Context, _ *proto.Empty) (*proto.VersionNumber, error) {
	v, err := s.Impl.GetVersion()
	return &proto.VersionNumber{Version: v}, err
}

// DatasourceBase is a rough implementation of the Datasource interface
// It implements the PushVersion and GetVersion functions for convenience
type DatasourceBase struct {
	version               string
	laniVersionConstraint version.Constraints
	laniVersion           string
}

// SetPluginVersion sets the plugin version string
func (b *DatasourceBase) SetPluginVersion(verStr string) {
	b.version = verStr
}

// SetVersionConstraints sets the version constraints on Laniakea
func (b *DatasourceBase) SetVersionConstraints(verStr string) error {
	constraints, err := version.NewConstraint(verStr)
	if err != nil {
		return err
	}
	b.laniVersionConstraint = constraints
	return nil
}

// GetLaniVersion returns the version of laniakea
func (b *DatasourceBase) GetLaniVersion() string {
	return b.laniVersion
}

// GetVersion returns the current plugin version if it has been set
func (b *DatasourceBase) GetVersion() (string, error) {
	if b.version == "" {
		return "", ErrPluginVersionNotSet
	}
	return b.version, nil
}

// PushVersion sets the laniakea version atrribute
func (b *DatasourceBase) PushVersion(versionNumber string) error {
	laniV, err := version.NewVersion(versionNumber)
	if err != nil {
		return err
	}
	if !b.laniVersionConstraint.Check(laniV) {
		return ErrLaniakeaVersionMismatch
	}
	b.laniVersion = versionNumber
	return nil
}
