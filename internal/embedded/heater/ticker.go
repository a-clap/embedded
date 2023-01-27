package heater

import (
	"time"
)

type timeTicker struct {
	*time.Ticker
}

var _ Ticker = (*timeTicker)(nil)

func newTimeTicker() *timeTicker {
	t := &timeTicker{Ticker: time.NewTicker(1 * time.Hour)}
	t.Ticker.Stop()
	return t
}

func (t *timeTicker) Start(d time.Duration) {
	t.Ticker.Reset(d)
}

func (t *timeTicker) Stop() {
	t.Ticker.Stop()
}

func (t *timeTicker) Tick() <-chan time.Time {
	return t.Ticker.C

}
