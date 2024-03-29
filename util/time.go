package util

import (
	t "time"
)

type timer struct{}

type Timer interface {
	Now() t.Time
	Tick(d t.Duration) <-chan t.Time
}

func NewTimer() *timer {
	return &timer{}
}

func (ti *timer) Now() t.Time {
	return t.Now()
}

func (ti *timer) Tick(d t.Duration) <-chan t.Time {
	return t.NewTicker(d).C
}

type testTimer struct {
	now  t.Time
	tick <-chan t.Time
}

func TestTimerFromUnix(ti int64) *testTimer {
	return &testTimer{
		now: t.Unix(ti, 0),
	}
}

func (tti *testTimer) Now() t.Time {
	return tti.now
}

func (tti *testTimer) Tick(t.Duration) <-chan t.Time {
	return tti.tick
}
