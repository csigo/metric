package metric

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	nowTS = time.Now().UnixNano()
)

func init() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			atomic.StoreInt64(&nowTS, now.UnixNano())
		}
	}()
}

// timeNow returns unix timestamp in nano-seconds
var timeNow = func() int64 {
	return atomic.LoadInt64(&nowTS)
}

// check checks range of window and bucket duration
func check(window, bucket time.Duration) error {
	switch {
	case window < time.Minute:
		return fmt.Errorf("invalid window duration less than minute %d", window)
	// Check if bucket size is less than one millisecond
	case bucket < 2*time.Second:
		return fmt.Errorf("invalid bucket duration less than 2 second %d", bucket)
	// Check if bukect size is multiple of millisecond
	case bucket%time.Second != 0:
		return fmt.Errorf("invalid bucket duration not multiple of second %d", bucket)
	// Check if bucket size is larger than window size
	case window < bucket:
		return fmt.Errorf("invalid pair window less than bucket %d:%d", window, bucket)
	// Check if window is divisible by window
	case window%bucket != 0:
		return fmt.Errorf("indivisible window bucket setting %d:%d", window, bucket)
	}
	return nil
}
