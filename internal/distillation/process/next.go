/*
 * Copyright (c) 2023 a-clap. All rights reserved.
 * Use of this source code is governed by a MIT-style license that can be found in the LICENSE file.
 */

package process

type MoveToNextType int

const (
	ByTime        MoveToNextType = iota
	ByTemperature MoveToNextType = iota
)

type moveToNext interface {
	next() bool
}

var (
	_ moveToNext = (*byTime)(nil)
	_ moveToNext = (*byTemperature)(nil)
)

type byTime struct {
	clock    Clock
	start    int64
	duration int64
}

func newByTime(clock Clock, duration int64) *byTime {
	return &byTime{
		clock:    clock,
		start:    clock.Unix(),
		duration: duration,
	}
}
func (b *byTime) next() bool {
	return (b.start + b.duration) < b.clock.Unix()
}

func (b *byTime) reset() {
	b.start = b.clock.Unix()
}

type byTemperature struct {
	waiting   bool
	byTime    *byTime
	threshold float64
	sensor    Sensor
	duration  int64
}

func newByTemperature(clock Clock, sensor Sensor, threshold float64, duration int64) *byTemperature {
	return &byTemperature{
		waiting:   false,
		byTime:    newByTime(clock, duration),
		threshold: threshold,
		sensor:    sensor,
		duration:  duration,
	}

}
func (b *byTemperature) next() bool {
	overThreshold := b.sensor.Temperature() > b.threshold
	// Did anything change since last call?
	if overThreshold != b.waiting {
		if overThreshold {
			// Okay, we are now over threshold
			// Start waiting for time
			b.waiting = true
			b.byTime.reset()
		} else {
			// Sadly, we are below threshold
			b.waiting = false
			return false
		}
	}
	return overThreshold && b.byTime.next()
}
