package processor

import (
	"context"
	"io"
	"sync"
	"time"
)

type (
	CloserFunc        func() error
	CloserContextFunc = func(ctx context.Context) error
)

func (cf CloserFunc) Close() error {
	return cf()
}

func NewCloserContextFunc(
	f CloserContextFunc,
	ctx context.Context, timeout time.Duration,
) CloserFunc {
	return func() error {
		if timeout > 0 {
			ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)

			defer cancel()

			return f(ctxWithTimeout)
		}

		return f(ctx)
	}
}

func WatchForShutdown(ctx context.Context, wg *sync.WaitGroup, closer io.Closer) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		_ = closer.Close()
	}()
}

func Wrap(ctx context.Context, wg *sync.WaitGroup, cb func(context.Context)) {
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		defer func() {
			if wg != nil {
				wg.Done()
			}
		}()

		select {
		case <-ctx.Done():
			return
		default:
			cb(ctx)
		}
	}()
}
