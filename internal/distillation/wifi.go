/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package distillation

type Wifi interface {
	APs() ([]string, error)
	Connect(ap, psk string) error
	Disconnect() error
	Status() (connected bool, ap string, err error)
}

type WifiHandler struct {
}
