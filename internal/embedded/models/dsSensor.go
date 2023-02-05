package models

type DSResolution int
type OnewireBusName string

const (
	Resolution9BIT  DSResolution = 9
	Resolution10BIT DSResolution = 10
	Resolution11BIT DSResolution = 11
	Resolution12BIT DSResolution = 12
	DefaultSamples  uint         = 5
)

type OnewireSensors struct {
	Bus      OnewireBusName `json:"bus"`
	DSConfig []DSConfig     `json:"ds18b20"`
}
type DSSensor interface {
	Temperature() Temperature
	Config() DSConfig
	SetConfig(cfg DSConfig) (err error)
}

type DSConfig struct {
	ID             string       `json:"id"`
	Enabled        bool         `json:"enabled"`
	Resolution     DSResolution `json:"resolution"`
	PollTimeMillis uint         `json:"poll_time_millis"`
	Samples        uint         `json:"samples"`
}
