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

type HeaterClient struct {
	addr    string
	timeout time.Duration
}

func NewHeaterClient(addr string, timeout time.Duration) *HeaterClient {
	return &HeaterClient{addr: addr, timeout: timeout}
}

func (p *HeaterClient) Get() ([]HeaterConfig, error) {
	return restclient.Get[[]HeaterConfig, *Error](p.addr+RoutesGetHeaters, p.timeout)
}

func (p *HeaterClient) Configure(setConfig HeaterConfig) (HeaterConfig, error) {
	return restclient.Put[HeaterConfig, *Error](p.addr+RoutesConfigHeater, p.timeout, setConfig)
}

type HeaterRPCClient struct {
	timeout time.Duration
	conn    *grpc.ClientConn
	client  embeddedproto.HeaterClient
}

func NewHeaterRPCCLient(addr string, timeout time.Duration) (*HeaterRPCClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &HeaterRPCClient{timeout: timeout, conn: conn, client: embeddedproto.NewHeaterClient(conn)}, nil
}

func (g *HeaterRPCClient) Get() ([]HeaterConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	got, err := g.client.HeaterGet(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	confs := make([]HeaterConfig, len(got.Configs))
	for i, elem := range got.Configs {
		confs[i] = rpcToHeaterConfig(elem)
	}
	return confs, nil
}

func (g *HeaterRPCClient) Configure(setConfig HeaterConfig) (HeaterConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	set := heaterConfigToRPC(&setConfig)
	got, err := g.client.HeaterConfigure(ctx, set)
	if err != nil {
		return HeaterConfig{}, err
	}
	setConfig = rpcToHeaterConfig(got)
	return setConfig, nil
}

func (g *HeaterRPCClient) Close() {
	_ = g.conn.Close()
}
