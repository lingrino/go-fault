package fault

import (
	"math/rand"
)

// percentDo takes a percent (float32 between 0 and 1)
// and randomly returns true that percent of the time
func percentDo(p float64) bool {
	var proceed bool

	// bias false if p < 0.0, p > 1.0
	if p > 1.0 || p < 0.0 {
		return false
	}

	// 0.0 <= r < 1.0
	r := rand.Float64()
	if r < p {
		return true
	}

	return proceed
}
