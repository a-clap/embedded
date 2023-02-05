/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

type rtd struct {
	r   uint16
	err error
}

func newRtd() *rtd {
	return &rtd{}
}

func (r *rtd) update(msb byte, lsb byte) error {
	// first bit in lsb is information about error
	if lsb&0x1 == 0x1 {
		r.err = ErrRtd
		return r.err
	}
	r.r = uint16(msb)<<8 | uint16(lsb)
	// rtd need to be shifted
	r.r >>= 1
	r.err = nil

	return nil
}

func (r *rtd) rtd() uint16 {
	return r.r
}
