/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package main

import (
	"github.com/a-clap/iot/internal/rest"
	"github.com/a-clap/iot/internal/wifi"
)

type WifiHandler struct {
	*wifi.Wifi
}

var _ rest.WIFIHandler = (*WifiHandler)(nil)

func New() (*WifiHandler, error) {
	w, err := wifi.New()
	if err != nil {
		return nil, err
	}
	return &WifiHandler{w}, nil
}

func (w *WifiHandler) APs() ([]rest.WIFINetwork, error) {
	aps, err := w.Wifi.APs()
	if err != nil {
		return nil, err
	}
	restAps := make([]rest.WIFINetwork, len(aps))
	for i, ap := range aps {
		restAps[i].SSID = ap.SSID
	}

	return restAps, nil
}

func (w *WifiHandler) Connect(n rest.WIFINetwork) error {
	if c, err := w.Connected(); err != nil {
		return err
	} else if c.Connected {
		_ = w.Disconnect()
	}
	net := wifi.Network{
		AP: wifi.AP{
			ID:   0,
			SSID: n.SSID,
		},
		Password: n.Password,
	}

	return w.Wifi.Connect(net)
}

func (w *WifiHandler) Status() (rest.WIFIStatus, error) {
	c, err := w.Wifi.Connected()
	if err != nil {
		return rest.WIFIStatus{}, err
	}
	return rest.WIFIStatus{
		Connected: c.Connected,
		SSID:      c.SSID,
	}, nil
}
