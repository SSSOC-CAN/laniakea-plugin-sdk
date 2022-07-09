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
	"github.com/hashicorp/go-version"
)

type ControllerGRPCClient struct{ client proto.ControllerClient }

type ControllerGRPCServer struct {
	proto.UnimplementedControllerServer
	Impl Controller
}

// Stop implements the Controller interface method Stop
func (c *ControllerGRPCClient) Stop() error {
	_, err := c.client.Stop(context.Background(), &proto.Empty{})
	if err != nil {
		return err
	}
	return nil
}

// Command implements the Controller interface method Command
func (c *ControllerGRPCClient) Command(f *proto.Frame) (chan *proto.Frame, error) {
	stream, err := c.client.Command(context.Background(), f)
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

// PushVersion implements the Controller interface method PushVersion
func (c *ControllerGRPCClient) PushVersion(versionNumber string) error {
	_, err := c.client.PushVersion(context.Background(), &proto.VersionNumber{Version: versionNumber})
	if err != nil {
		return err
	}
	return nil
}

// GetVersion implements the Controller interface method GetVersion
func (c *ControllerGRPCClient) GetVersion() (string, error) {
	resp, err := c.client.GetVersion(context.Background(), &proto.Empty{})
	if err != nil {
		return "", err
	}
	return resp.Version, nil
}

// Stop implements the Controller gRPC server interface
func (s *ControllerGRPCServer) Stop(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	err := s.Impl.Stop()
	return &proto.Empty{}, err
}

// Command implements the Controller gRPC server interface
func (s *ControllerGRPCServer) Command(req *proto.Frame, stream proto.Controller_CommandServer) error {
	frameChan, err := s.Impl.Command(req)
	if err != nil {
		return err
	}
	for {
		select {
		case frame := <-frameChan:
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

// PushVersion implements the Controller gRPC server interface
func (s *ControllerGRPCServer) PushVersion(ctx context.Context, req *proto.VersionNumber) (*proto.Empty, error) {
	err := s.Impl.PushVersion(req.Version)
	return &proto.Empty{}, err
}

// GetVersion implements the Controller gRPC server interface
func (s *ControllerGRPCServer) GetVersion(ctx context.Context, _ *proto.Empty) (*proto.VersionNumber, error) {
	v, err := s.Impl.GetVersion()
	return &proto.VersionNumber{Version: v}, err
}

// ControllerBase is a rough implementation of the Controller interface
// It implements the PushVersion and GetVersion functions for convenience
type ControllerBase struct {
	version               string
	laniVersionConstraint version.Constraints
	laniVersion           string
}

// SetPluginVersion sets the plugin version string
func (b *ControllerBase) SetPluginVersion(verStr string) {
	b.version = verStr
}

// SetVersionConstraints sets the version constraints on Laniakea
func (b *ControllerBase) SetRequiredVersion(verStr string) error {
	constraints, err := version.NewConstraint(verStr)
	if err != nil {
		return err
	}
	b.laniVersionConstraint = constraints
	return nil
}

// GetLaniVersion returns the version of laniakea
func (b *ControllerBase) GetLaniVersion() string {
	return b.laniVersion
}

// GetVersion returns the current plugin version if it has been set
func (b *ControllerBase) GetVersion() (string, error) {
	if b.version == "" {
		return "", ErrPluginVersionNotSet
	}
	return b.version, nil
}

// PushVersion sets the required laniakea version
func (b *ControllerBase) PushVersion(versionNumber string) error {
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
