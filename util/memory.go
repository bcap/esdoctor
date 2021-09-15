package util

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

func LogMemoryUsage(when string, level log.Level) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	toMb := func(x uint64) int64 {
		return int64(x / uint64(1024) / uint64(1024))
	}
	log.StandardLogger().Logf(
		level,
		"Mem usage %s: Alloc=%dMiB, TotalAlloc=%dMiB, GCRuns=%d",
		when, toMb(m.Alloc), toMb(m.TotalAlloc), m.NumGC,
	)
}
