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
	next(stamp int64) bool
	timeleft(stamp int64) int64
}

var (
	_ moveToNext = (*byTime)(nil)
	_ moveToNext = (*byTemperature)(nil)
)

type byTime struct {
	start    int64
	duration int64
}

func newByTime(stamp int64, duration int64) *byTime {
	return &byTime{
		start:    stamp,
		duration: duration,
	}
}
func (b *byTime) next(stamp int64) bool {
	return b.timeleft(stamp) == 0
}

func (b *byTime) reset(stamp int64) {
	b.start = stamp
}

func (b *byTime) timeleft(stamp int64) int64 {
	t := (b.start + b.duration) - stamp
	if t < 0 {
		return 0
	}
	return t
}

type byTemperature struct {
	waiting   bool
	byTime    *byTime
	threshold float64
	sensor    Sensor
	duration  int64
}

func (b *byTemperature) timeleft(stamp int64) int64 {
	if b.waiting {
		return b.byTime.timeleft(stamp)
	}
	return b.duration
}

func newByTemperature(stamp int64, sensor Sensor, threshold float64, duration int64) *byTemperature {
	return &byTemperature{
		waiting:   false,
		byTime:    newByTime(stamp, duration),
		threshold: threshold,
		sensor:    sensor,
		duration:  duration,
	}

}
func (b *byTemperature) next(stamp int64) bool {
	overThreshold := b.sensor.Temperature() > b.threshold
	// Did anything change since last call?
	if overThreshold != b.waiting {
		if overThreshold {
			// Okay, we are now over threshold
			// Start waiting for time
			b.waiting = true
			b.byTime.reset(stamp)
		} else {
			// Sadly, we are below threshold
			b.waiting = false
			return false
		}
	}
	return overThreshold && b.byTime.next(stamp)
}
