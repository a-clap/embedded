/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import (
	"context"
	"time"
	
	"github.com/a-clap/embedded/pkg/embedded/embeddedproto"
	"github.com/a-clap/embedded/pkg/restclient"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PTClient struct {
	addr    string
	timeout time.Duration
}

func NewPTClient(addr string, timeout time.Duration) *PTClient {
	return &PTClient{addr: addr, timeout: timeout}
}

func (p *PTClient) Get() ([]PTSensorConfig, error) {
	return restclient.Get[[]PTSensorConfig, *Error](p.addr+RoutesGetPT100Sensors, p.timeout)
}

func (p *PTClient) Configure(setConfig PTSensorConfig) (PTSensorConfig, error) {
	return restclient.Put[PTSensorConfig, *Error](p.addr+RoutesConfigPT100Sensor, p.timeout, setConfig)
}

func (p *PTClient) Temperatures() ([]PTTemperature, error) {
	return restclient.Get[[]PTTemperature, *Error](p.addr+RoutesGetPT100Temperatures, p.timeout)
}

type PTRPCClient struct {
	timeout time.Duration
	conn    *grpc.ClientConn
	client  embeddedproto.PTClient
}

func NewPTRPCClient(addr string, timeout time.Duration) (*PTRPCClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &PTRPCClient{timeout: timeout, conn: conn, client: embeddedproto.NewPTClient(conn)}, nil
}

func (g *PTRPCClient) Get() ([]PTSensorConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	got, err := g.client.PTGet(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	confs := make([]PTSensorConfig, len(got.Configs))
	for i, elem := range got.Configs {
		confs[i] = rpcToPTConfig(elem)
	}
	return confs, nil
}

func (g *PTRPCClient) Configure(setConfig PTSensorConfig) (PTSensorConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	set := ptConfigToRPC(&setConfig)
	got, err := g.client.PTConfigure(ctx, set)
	if err != nil {
		return PTSensorConfig{}, err
	}
	setConfig = rpcToPTConfig(got)
	return setConfig, nil
}

func (g *PTRPCClient) Temperatures() ([]PTTemperature, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	got, err := g.client.PTGetTemperatures(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return rpcToPTTemperature(got), nil
}

func (g *PTRPCClient) Close() {
	_ = g.conn.Close()
}
