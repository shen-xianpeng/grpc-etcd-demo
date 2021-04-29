package installation

import (
	"context"
	"errors"
	"go-kit-etcd-demo/lib/logger"
	"go-kit-etcd-demo/lib/proto/installation"
)

type Server interface {
	Start() error
	Stop() error
}

type option func(Server) error

type InstallationServer struct {
	addr string
}

func NewInstallationServer(opts ...option) (*InstallationServer, []error) {
	s := &InstallationServer{}
	errs := make([]error, 0)

	// call option functions on instance to set options on it
	for _, opt := range opts {
		err := opt(s)
		// if the option func returns an error, add it to the list of errors
		if err != nil {
			errs = append(errs, err)
		}
	}

	return s, errs
}

func (i InstallationServer) Start() error {

	return nil
}

func (i InstallationServer) Stop() error {

	return nil
}

func Addr(addr string) option {
	return func(i Server) error {
		s, ok := i.(*InstallationServer)
		if ok == false {
			return errors.New(" need InstallationServer")
		}
		s.addr = addr
		return nil
	}
}

func (i *InstallationServer) ServerInfo(ctx context.Context, req *installation.ServerInfoRequest) (*installation.ServerInfoResponse, error) {
	logger.Info()
	return &installation.ServerInfoResponse{
		Addr: i.addr,
	}, nil
}

func (i *InstallationServer) RegisterDevice(ctx context.Context, req *installation.RegisterDeviceRequest) (*installation.RegisterDeviceResponse, error) {
	logger.InfoMsg(req.DeviceId)
	return &installation.RegisterDeviceResponse{
		DeviceId:     req.DeviceId,
		DeviceSecret: "Secret Key",
	}, nil
}
