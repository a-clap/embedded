/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package max31865

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_rtd(t *testing.T) {
	tests := []struct {
		name     string
		msb, lsb byte
		rtd      uint16
		wantErr  bool
		errType  error
	}{
		{
			name:    "detect error",
			msb:     0xff,
			lsb:     0x1,
			rtd:     0,
			wantErr: true,
			errType: ErrRtd,
		},
		{
			name:    "rtd_1",
			msb:     0xff,
			lsb:     0xfe,
			rtd:     32767,
			wantErr: false,
			errType: nil,
		},
		{
			name:    "rtd_2",
			msb:     0x43,
			lsb:     0x38,
			rtd:     8604,
			wantErr: false,
			errType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rtd := new(rtd)
			err := rtd.update(tt.msb, tt.lsb)
			if tt.wantErr {
				require.NotNil(t, err)
				require.ErrorIs(t, err, ErrRtd)
				return
			}
			require.Nil(t, err)
			require.EqualValues(t, tt.rtd, rtd.rtd())
		})
	}
}
