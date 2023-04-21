package embedded

import (
	"context"
	"net"
	
	"github.com/a-clap/embedded/pkg/embedded/embeddedproto"
	"github.com/a-clap/logging"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type RPC struct {
	embeddedproto.UnimplementedGPIOServer
	*Embedded
}

func NewRPC(options ...Option) (*RPC, error) {
	r := &RPC{}
	
	e, err := New(options...)
	if err != nil {
		return nil, err
	}
	e.url = ":50051"
	r.Embedded = e
	
	return r, nil
}

func (r *RPC) Run() error {
	listener, err := net.Listen("tcp", r.Embedded.url)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	embeddedproto.RegisterGPIOServer(s, r)
	return s.Serve(listener)
}

func (r *RPC) Close() {
	r.Embedded.close()
}
func (r *RPC) GPIOGet(ctx context.Context, empty *empty.Empty) (*embeddedproto.GPIOConfigs, error) {
	logger.Debug("GPIOGet")
	g, err := r.Embedded.GPIO.GetConfigAll()
	if err != nil {
		logger.Error("GPIOGet", logging.String("error", err.Error()))
		return nil, err
	}
	configs := make([]*embeddedproto.GPIOConfig, len(g))
	for i, elem := range g {
		configs[i] = gpioConfigToRPC(&elem)
	}
	return &embeddedproto.GPIOConfigs{Configs: configs}, nil
}

func (r *RPC) GPIOConfigure(c context.Context, cfg *embeddedproto.GPIOConfig) (*embeddedproto.GPIOConfig, error) {
	logger.Debug("GPIOConfigure")
	config := rpcToGPIOConfig(cfg)
	err := r.Embedded.GPIO.SetConfig(config)
	if err != nil {
		return nil, err
	}
	config, err = r.Embedded.GPIO.GetConfig(cfg.ID)
	if err != nil {
		return nil, err
	}
	cfg = gpioConfigToRPC(&config)
	return cfg, nil
}
