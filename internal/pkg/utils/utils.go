package utils

import (
	"time"
)

func Must[T any](val T, e error) T {
	if e != nil {
		panic(e)
	}

	return val
}

func Find[T any](s []T, pred func(T, int, []T) bool) T {
	var val T

	for i := 0; i < len(s); i++ {
		if pred(s[i], i, s) {
			return s[i]
		}
	}

	return val
}

func Debounce(d time.Duration, f func()) func() {
	var timer *time.Timer

	return func() {
		if timer != nil && !timer.Stop() {
			// drain the channel
			<-timer.C
		}

		timer = time.AfterFunc(d, func() {
			defer func() { timer = nil }()
			f()
		})
	}
}

func AsyncResult[T any](f func() T) <-chan T {
	r := make(chan T)

	go func() {
		defer close(r)
		r <- f()
	}()

	return r
}

// CloseSafely closes buffered or unbuffered channel if it is not already closed.
func CloseSafely[T any](c chan T) {
	defer recover()
	close(c)
}
