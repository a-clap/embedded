package models

type WifiStatus struct {
	Connected bool   `json:"connected"`
	SSID      string `json:"ssid"`
}

type WifiNetwork struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
}
