package models

type PTHandler interface {
	ID() string

	ASyncPoll(data chan PollData) error
	SyncPoll(data chan PollData) error

	StopPoll() error
}

type PTConfig struct {
}
