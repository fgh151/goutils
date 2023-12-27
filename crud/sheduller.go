package crud

import (
	"context"
	"time"
)

func Schedule(ctx context.Context, p time.Duration, f func(time time.Time)) {
	t := time.NewTicker(p)
	for {
		select {
		case v := <-t.C:
			f(v)
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}
