package policy

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter bounds the load one model can put on a cluster: one call at a time,
// and no more than so many a minute.
type Limiter struct {
	perMinute int
	slot      chan struct{}

	mu    sync.Mutex
	calls []time.Time
	now   func() time.Time
}

// NewLimiter allows perMinute calls a minute, one at a time.
func NewLimiter(perMinute int) *Limiter {
	return &Limiter{perMinute: perMinute, slot: make(chan struct{}, 1), now: time.Now}
}

// Acquire waits for the call before to finish, and refuses a call past the
// rate. The returned release has to be called when the call is done.
func (l *Limiter) Acquire(ctx context.Context) (release func(), err error) {
	if err := l.count(); err != nil {
		return nil, err
	}
	select {
	case l.slot <- struct{}{}:
		return func() { <-l.slot }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// count records a call, or refuses it when the minute's are used up.
func (l *Limiter) count() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-time.Minute)
	kept := l.calls[:0]
	for _, t := range l.calls {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.calls = kept

	if len(l.calls) >= l.perMinute {
		wait := l.calls[0].Add(time.Minute).Sub(now).Round(time.Second)
		return refused(fmt.Sprintf("more than %d calls a minute; try again in %s", l.perMinute, wait))
	}
	l.calls = append(l.calls, now)
	return nil
}
