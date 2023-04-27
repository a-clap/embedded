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
	url string
	embeddedproto.UnimplementedPTServer
	embeddedproto.UnimplementedHeaterServer
	embeddedproto.UnimplementedDSServer
	embeddedproto.UnimplementedGPIOServer
	*Embedded
}

func NewRPC(url string, options ...Option) (*RPC, error) {
	r := &RPC{
		url: url,
	}

	e, err := New(options...)
	if err != nil {
		return nil, err
	}
	r.Embedded = e

	return r, nil
}

func (r *RPC) Run() error {
	listener, err := net.Listen("tcp", r.url)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	embeddedproto.RegisterGPIOServer(s, r)
	embeddedproto.RegisterDSServer(s, r)
	embeddedproto.RegisterPTServer(s, r)
	embeddedproto.RegisterHeaterServer(s, r)

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

func (r *RPC) DSGet(ctx context.Context, e *empty.Empty) (*embeddedproto.DSConfigs, error) {
	logger.Debug("DSGet")
	g := r.Embedded.DS.GetSensors()

	configs := make([]*embeddedproto.DSConfig, len(g))
	for i, elem := range g {
		configs[i] = dsConfigToRPC(&elem)
	}
	return &embeddedproto.DSConfigs{Configs: configs}, nil
}

func (r *RPC) DSConfigure(ctx context.Context, config *embeddedproto.DSConfig) (*embeddedproto.DSConfig, error) {
	logger.Debug("DSConfigure")
	cfg := rpcToDSConfig(config)
	newCfg, err := r.Embedded.DS.SetConfig(cfg)
	if err != nil {
		return nil, err
	}
	return dsConfigToRPC(&newCfg), nil
}

func (r *RPC) DSGetTemperatures(ctx context.Context, e *empty.Empty) (*embeddedproto.DSTemperatures, error) {
	logger.Debug("DSGetTemperatures")
	t := r.Embedded.DS.GetTemperatures()
	return dsTemperatureToRPC(t), nil
}

func (r *RPC) PTGet(ctx context.Context, e *empty.Empty) (*embeddedproto.PTConfigs, error) {
	logger.Debug("PTGet")
	g := r.Embedded.PT.GetSensors()

	configs := make([]*embeddedproto.PTConfig, len(g))
	for i, elem := range g {
		configs[i] = ptConfigToRPC(&elem)
	}
	return &embeddedproto.PTConfigs{Configs: configs}, nil
}

func (r *RPC) PTConfigure(ctx context.Context, config *embeddedproto.PTConfig) (*embeddedproto.PTConfig, error) {
	logger.Debug("PTConfigure")
	cfg := rpcToPTConfig(config)
	newCfg, err := r.Embedded.PT.SetConfig(cfg)
	if err != nil {
		return nil, err
	}
	return ptConfigToRPC(&newCfg), nil
}

func (r *RPC) PTGetTemperatures(ctx context.Context, e *empty.Empty) (*embeddedproto.PTTemperatures, error) {
	logger.Debug("PTGetTemperatures")
	t := r.Embedded.PT.GetTemperatures()
	return ptTemperatureToRPC(t), nil
}

func (r *RPC) HeaterGet(context.Context, *empty.Empty) (*embeddedproto.HeaterConfigs, error) {
	logger.Debug("HeaterGet")
	g := r.Embedded.Heaters.Get()

	configs := make([]*embeddedproto.HeaterConfig, len(g))
	for i, elem := range g {
		configs[i] = heaterConfigToRPC(&elem)
	}
	return &embeddedproto.HeaterConfigs{Configs: configs}, nil
}

func (r *RPC) HeaterConfigure(ctx context.Context, config *embeddedproto.HeaterConfig) (*embeddedproto.HeaterConfig, error) {
	logger.Debug("HeaterConfigure")
	cfg := rpcToHeaterConfig(config)
	err := r.Embedded.Heaters.SetConfig(cfg)
	if err != nil {
		return nil, err
	}
	newCfg, err := r.Embedded.Heaters.ConfigBy(config.ID)
	if err != nil {
		return nil, err
	}

	return heaterConfigToRPC(&newCfg), nil
}
