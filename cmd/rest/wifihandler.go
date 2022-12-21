package main

import (
	"github.com/a-clap/iot/internal/models"
	"github.com/a-clap/iot/internal/rest"
	"github.com/a-clap/iot/internal/wifi"
)

type WifiHandler struct {
	*wifi.Wifi
}

var _ rest.WifiHandler = (*WifiHandler)(nil)

func New() (*WifiHandler, error) {
	w, err := wifi.New()
	if err != nil {
		return nil, err
	}
	return &WifiHandler{w}, nil
}

func (w *WifiHandler) APs() ([]models.WifiNetwork, error) {
	aps, err := w.Wifi.APs()
	if err != nil {
		return nil, err
	}
	restAps := make([]models.WifiNetwork, len(aps))
	for i, ap := range aps {
		restAps[i].SSID = ap.SSID
	}

	return restAps, nil
}

func (w *WifiHandler) Connect(n models.WifiNetwork) error {
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

func (w *WifiHandler) Status() (models.WifiStatus, error) {
	c, err := w.Wifi.Connected()
	if err != nil {
		return models.WifiStatus{}, err
	}
	return models.WifiStatus{
		Connected: c.Connected,
		SSID:      c.SSID,
	}, nil
}
