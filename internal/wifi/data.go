/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

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
