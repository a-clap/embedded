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

type DS18B20Client struct {
	addr    string
	timeout time.Duration
}

func NewDS18B20Client(addr string, timeout time.Duration) *DS18B20Client {
	return &DS18B20Client{addr: addr, timeout: timeout}
}

func (p *DS18B20Client) Get() ([]DSSensorConfig, error) {
	return restclient.Get[[]DSSensorConfig, *Error](p.addr+RoutesGetOnewireSensors, p.timeout)
}

func (p *DS18B20Client) Configure(setConfig DSSensorConfig) (DSSensorConfig, error) {
	return restclient.Put[DSSensorConfig, *Error](p.addr+RoutesConfigOnewireSensor, p.timeout, setConfig)
}

func (p *DS18B20Client) Temperatures() ([]DSTemperature, error) {
	return restclient.Get[[]DSTemperature, *Error](p.addr+RoutesGetOnewireTemperatures, p.timeout)
}

type DSRPCClient struct {
	timeout time.Duration
	conn    *grpc.ClientConn
	client  embeddedproto.DSClient
}

func NewDSRPCClient(addr string, timeout time.Duration) (*DSRPCClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &DSRPCClient{timeout: timeout, conn: conn, client: embeddedproto.NewDSClient(conn)}, nil
}

func (g *DSRPCClient) Get() ([]DSSensorConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	got, err := g.client.DSGet(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	confs := make([]DSSensorConfig, len(got.Configs))
	for i, elem := range got.Configs {
		confs[i] = rpcToDSConfig(elem)
	}
	return confs, nil
}

func (g *DSRPCClient) Configure(setConfig DSSensorConfig) (DSSensorConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	set := dsConfigToRPC(&setConfig)
	got, err := g.client.DSConfigure(ctx, set)
	if err != nil {
		return DSSensorConfig{}, err
	}
	setConfig = rpcToDSConfig(got)
	return setConfig, nil
}

func (g *DSRPCClient) Temperatures() ([]DSTemperature, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	got, err := g.client.DSGetTemperatures(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return rpcToDSTemperature(got), nil
}

func (g *DSRPCClient) Close() {
	_ = g.conn.Close()
}
