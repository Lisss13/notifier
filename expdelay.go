package notifier

import "time"

// ExpDelay implements exponential delay and provides a convenient way to exponential delay..
type ExpDelay struct{ cur, max time.Duration }

// New returns new exponential delay which start with min delay, increase
// each next delay in 2 times up to max delay.
//
//	for delay := expdelay.New(minDelay, maxDelay); ; delay.Sleep() {
//		err := op()
//		if err == nil {
//			break
//		}
//	}
func NewExpDelay(min, max time.Duration) *ExpDelay {
	return &ExpDelay{
		cur: min,
		max: max,
	}
}

// Sleep will call time.Sleep using current delay.
func (d *ExpDelay) Sleep() {
	time.Sleep(d.cur)
	d.cur *= 2
	if d.cur > d.max {
		d.cur = d.max
	}
}
