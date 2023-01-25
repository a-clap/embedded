package avg_test

import (
	"github.com/a-clap/iot/pkg/avg"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AvgTestSuite struct {
	suite.Suite
}

func TestAvgTestSuite(t *testing.T) {
	suite.Run(t, new(AvgTestSuite))
}

func (t *AvgTestSuite) TestHandleErrors() {
	a, err := avg.New[int](0)
	t.Nil(a)
	t.NotNil(err)
	t.ErrorIs(avg.ErrSizeIsZero, err)

	a, err = avg.New[int](2)
	t.NotNil(a)
	t.Nil(err)

	err = a.Resize(0)
	t.ErrorIs(avg.ErrSizeIsZero, err)

	err = a.Resize(2)
	t.Nil(err)

}

func (t *AvgTestSuite) TestResize() {
	args := []struct {
		name                                            string
		initSize, newSize                               uint
		initValues, nextValues                          []int
		initExpected, afterResizeExpected, nextExpected int
	}{
		{
			name:                "make buffer bigger (values alredy present)",
			initSize:            2,
			newSize:             5,
			initValues:          []int{2, 2},
			nextValues:          []int{104, 104, 103},
			initExpected:        2,
			afterResizeExpected: 2,
			nextExpected:        63,
		},
		{
			name:                "make buffer bigger (no values inside)",
			initSize:            2,
			newSize:             5,
			initValues:          []int{2, 2},
			nextValues:          []int{104, 104, 103},
			initExpected:        2,
			afterResizeExpected: 2,
			nextExpected:        63,
		},
		{
			name:                "make buffer smaller (no values)",
			initSize:            5,
			newSize:             2,
			initValues:          []int{},
			nextValues:          []int{510},
			initExpected:        0,
			afterResizeExpected: 0,
			nextExpected:        510,
		},
		{
			name:                "make buffer bigger (no values)",
			initSize:            2,
			newSize:             5,
			initValues:          []int{},
			nextValues:          []int{510},
			initExpected:        0,
			afterResizeExpected: 0,
			nextExpected:        510,
		},
		{
			name:                "make buffer single (values already present)",
			initSize:            5,
			newSize:             1,
			initValues:          []int{101, 102, 103, 104, 105},
			nextValues:          []int{523},
			initExpected:        103,
			afterResizeExpected: 105,
			nextExpected:        523,
		},
		{
			name:                "grow from single",
			initSize:            1,
			newSize:             5,
			initValues:          []int{105},
			nextValues:          []int{101, 102, 103, 104},
			initExpected:        105,
			afterResizeExpected: 105,
			nextExpected:        103,
		},
	}
	for _, arg := range args {
		a, _ := avg.New[int](arg.initSize)
		for _, initV := range arg.initValues {
			a.Add(initV)
		}
		t.EqualValues(arg.initExpected, a.Average(), arg.name)

		_ = a.Resize(arg.newSize)

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
		a, _ := avg.New[float64](arg.size)
		for _, elem := range arg.values {
			a.Add(elem)
		}
		t.InDelta(arg.expected, a.Average(), 0.01, arg.name)

	}

}

func (t *AvgTestSuite) TestAverage_Int() {

	args := []struct {
		name     string
		size     uint
		values   []int
		expected int
	}{
		{
			name:     "basic",
			size:     3,
			values:   []int{1, 2, 3},
			expected: 2,
		},
		{
			name:     "less elements than size",
			size:     3,
			values:   []int{1, 9},
			expected: 5,
		},
		{
			name:     "more elements than size",
			size:     3,
			values:   []int{3, 3, 3, 6},
			expected: 4,
		},
		{
			name:     "much more elements than size",
			size:     2,
			values:   []int{0, 0, 0, 6, 12, 17, 34, 56, 100, 100},
			expected: 100,
		},
		{
			name:     "none elements",
			size:     2,
			values:   []int{},
			expected: 0,
		},
	}
	for _, arg := range args {
		a, _ := avg.New[int](arg.size)
		for _, elem := range arg.values {
			a.Add(elem)
		}
		t.EqualValues(arg.expected, a.Average())

	}

}
