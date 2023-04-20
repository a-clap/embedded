/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package avg_test

import (
	"testing"

	"github.com/a-clap/embedded/pkg/avg"
	"github.com/stretchr/testify/suite"
)

type AvgTestSuite struct {
	suite.Suite
}

func TestAvgTestSuite(t *testing.T) {
	suite.Run(t, new(AvgTestSuite))
}

func (t *AvgTestSuite) TestNewAndResize() {
	args := []struct {
		name                                            string
		initSize, newSize                               uint
		initValues, nextValues                          []float64
		initExpected, afterResizeExpected, nextExpected int
	}{
		{
			name:                "start with 0 as size",
			initSize:            0,
			newSize:             2,
			initValues:          []float64{2},
			nextValues:          []float64{8, 8},
			initExpected:        2,
			afterResizeExpected: 2,
			nextExpected:        8,
		},
		{
			name:                "make buffer bigger (values already present)",
			initSize:            2,
			newSize:             5,
			initValues:          []float64{2, 2},
			nextValues:          []float64{104, 104, 103},
			initExpected:        2,
			afterResizeExpected: 2,
			nextExpected:        63,
		},
		{
			name:                "make buffer bigger (no values inside)",
			initSize:            2,
			newSize:             5,
			initValues:          []float64{2, 2},
			nextValues:          []float64{104, 104, 103},
			initExpected:        2,
			afterResizeExpected: 2,
			nextExpected:        63,
		},
		{
			name:                "make buffer smaller (no values)",
			initSize:            5,
			newSize:             2,
			initValues:          []float64{},
			nextValues:          []float64{510},
			initExpected:        0,
			afterResizeExpected: 0,
			nextExpected:        510,
		},
		{
			name:                "make buffer bigger (no values)",
			initSize:            2,
			newSize:             5,
			initValues:          []float64{},
			nextValues:          []float64{510},
			initExpected:        0,
			afterResizeExpected: 0,
			nextExpected:        510,
		},
		{
			name:                "make buffer single (values already present)",
			initSize:            5,
			newSize:             1,
			initValues:          []float64{101, 102, 103, 104, 105},
			nextValues:          []float64{523},
			initExpected:        103,
			afterResizeExpected: 105,
			nextExpected:        523,
		},
		{
			name:                "grow from single",
			initSize:            1,
			newSize:             5,
			initValues:          []float64{105},
			nextValues:          []float64{101, 102, 103, 104},
			initExpected:        105,
			afterResizeExpected: 105,
			nextExpected:        103,
		},
		{
			name:                "resize to 0",
			initSize:            5,
			newSize:             0,
			initValues:          []float64{205, 200, 200, 200, 200},
			nextValues:          []float64{101},
			initExpected:        201,
			afterResizeExpected: 200,
			nextExpected:        101,
		},
	}
	for _, arg := range args {
		a := avg.New(arg.initSize)
		for _, initV := range arg.initValues {
			a.Add(initV)
		}
		t.EqualValues(arg.initExpected, a.Average(), arg.name)

		a.Resize(arg.newSize)

		t.EqualValues(arg.afterResizeExpected, a.Average(), arg.name)

		for _, next := range arg.nextValues {
			a.Add(next)
		}
		t.EqualValues(arg.nextExpected, a.Average(), arg.name)
	}

}

func (t *AvgTestSuite) TestAverage_Float() {

	args := []struct {
		name     string
		size     uint
		values   []float64
		expected float64
	}{
		{
			name:     "basic",
			size:     3,
			values:   []float64{1.5, 2.5, 4},
			expected: 2.66,
		},
		{
			name:     "less elements than size",
			size:     3,
			values:   []float64{1.11, 9.123},
			expected: 5.1165,
		},
		{
			name:     "more elements than size",
			size:     3,
			values:   []float64{1000.0, 3.789, 3.123, 6.6},
			expected: 4.504,
		},
		{
			name:     "much more elements than size",
			size:     2,
			values:   []float64{0, 0, 0, 6, 12, 17, 34, 56, 123.456, 789.123},
			expected: 456.2895,
		},
	}
	for _, arg := range args {
		a := avg.New(arg.size)
		for _, elem := range arg.values {
			a.Add(elem)
		}
		t.InDelta(arg.expected, a.Average(), 0.01, arg.name)

	}
}
