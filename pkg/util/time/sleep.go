package time

import (
	"math/rand"
	"time"
)

func SleepUpTo(duration time.Duration) {
	// #nosec G404 - math rand is enough
	time.Sleep(time.Duration(rand.Intn(int(duration.Milliseconds()))) * time.Millisecond)
}
