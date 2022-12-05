package wifi

type Event struct {
	ID      ID
	Message string
}

type ID int

const (
	Connected ID = iota
	Disconnected
	NetworkNotFound
	AuthReject
	OtherError
	Any
)

// AP stands for access point
type AP struct {
	ID   int
	SSID string
}

// Network is AP with provided password
type Network struct {
	AP
	Password string
}

// Status hold current status of wireless client
type Status struct {
	Connected bool
	SSID      string
}
