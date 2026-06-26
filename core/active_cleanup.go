package core

import (
	"log"
	"time"
)

func expireSample() float64 {
	const SAMPLE_SIZE int = 20 // Based on redis source code, we can sample 20 keys for expiration check
	var limit int = SAMPLE_SIZE
	var expiredCount int = 0

	for key := range store {
		if store[key].ExpiresAt != -1 {
			limit--
			if time.Now().UnixMilli() > store[key].ExpiresAt {
				Delete(key)
				expiredCount++
			}
		}

		if limit == 0 {
			break
		}
	}

	return float64(expiredCount) / float64(SAMPLE_SIZE)
}

func ActiveCleanup() {
	start := time.Now()
	for {
		expiredRatio := expireSample()
		if expiredRatio <= 0.25 {
			break
		}
		if time.Since(start) > 25*time.Millisecond {
			break
		}
	}
	log.Println("Active cleanup completed. Total keys remaining:", len(store))
}
