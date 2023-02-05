/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package rest

import (
	"encoding/json"
)

type Error struct {
	ErrorCode int    `json:"error_code"`
	Desc      string `json:"description"`
}

var _ error = Error{}

func (e Error) Error() string {
	return e.Desc
}

func (e Error) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func errorInterface(e error) Error {
	err := ErrInterface
	err.Desc = e.Error()
	return err
}

const (
	_ = -iota
	NotImplemented
	NotFound
	Interface
)

var (
	ErrNotImplemented = Error{ErrorCode: NotImplemented, Desc: "not implemented"}
	ErrNotFound       = Error{ErrorCode: NotFound, Desc: "not found"}
	ErrInterface      = Error{ErrorCode: Interface, Desc: ""}
)
