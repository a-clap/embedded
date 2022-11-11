package rest

import (
	"encoding/json"
)

type Error struct {
	ErrorCode int    `json:"error_code"`
	Desc      string `json:"description"`
}

var _ error = &Error{}

func (e *Error) Error() string {
	return e.Desc
}

func (e *Error) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

const (
	_ = -iota
	NotImplemented
	NotFound
)

var (
	ErrNotImplemented = Error{ErrorCode: NotImplemented, Desc: "not implemented"}
	ErrNotFound       = Error{ErrorCode: NotFound, Desc: "not found"}
)
