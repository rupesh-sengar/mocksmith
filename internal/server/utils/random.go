package utils

import "time"

func Roll(pct int) bool {
	if pct <= 0 {
		return false
	}
	if pct >= 100 {
		return true
	}
	var seed = time.Now().UnixNano()
	seed = (1103515245*seed + 12345) & 0x7fffffff
	return int(seed%100) < pct
}
