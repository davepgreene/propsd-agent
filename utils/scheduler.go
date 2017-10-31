package utils

import "time"

func Schedule(f func(), delay time.Duration) {
	t := time.NewTicker(delay)

	go func() {
		for range t.C {
			f()
		}
	}()
}
