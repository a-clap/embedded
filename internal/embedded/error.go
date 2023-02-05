/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package embedded

import "encoding/json"

type Error struct {
	Err string `json:"err"`
}

var _ error = Error{}

var (
	ErrHeaterDoesntExist = Error{Err: "heater doesn't exist"}
)

func (e Error) Error() string {
	return e.Err
}

func (e Error) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func toError(e error) Error {
	return Error{Err: e.Error()}
}
