package threadrepo

import (
	"log"
	"sync/atomic"
	"time"
)

// forYouMetrics records cache hit/miss + DB latency counters for the For You
// pipeline. All fields use sync/atomic so increments are lock-free. There is
// no external dashboard wiring (no Prometheus / OTel deps in this codebase
// yet); instead, key events emit a structured log line so that whatever log
// aggregator runs in prod (Loki, Cloudwatch, etc.) can compute aggregates.
//
// Snapshot() returns a flat struct safe to serialize into an admin-only
// `/admin/metrics/feed` endpoint when one is wired up.
type forYouMetrics struct {
	CacheHitOwn        atomic.Int64
	CacheMissOwn       atomic.Int64
	CacheHitFollowing  atomic.Int64
	CacheMissFollowing atomic.Int64
	CacheHitCommunity  atomic.Int64
	CacheMissCommunity atomic.Int64
	CacheHitGlobal     atomic.Int64
	CacheMissGlobal    atomic.Int64
	CacheHitNewCount   atomic.Int64
	CacheMissNewCount  atomic.Int64
	CacheHitSeed       atomic.Int64
	CacheMissSeed      atomic.Int64
}

// ForYouMetricsSnapshot is the externally-visible counter dump.
type ForYouMetricsSnapshot struct {
	CacheHitOwn        int64 `json:"cache_hit_own"`
	CacheMissOwn       int64 `json:"cache_miss_own"`
	CacheHitFollowing  int64 `json:"cache_hit_following"`
	CacheMissFollowing int64 `json:"cache_miss_following"`
	CacheHitCommunity  int64 `json:"cache_hit_community"`
	CacheMissCommunity int64 `json:"cache_miss_community"`
	CacheHitGlobal     int64 `json:"cache_hit_global"`
	CacheMissGlobal    int64 `json:"cache_miss_global"`
	CacheHitNewCount   int64 `json:"cache_hit_newcount"`
	CacheMissNewCount  int64 `json:"cache_miss_newcount"`
	CacheHitSeed       int64 `json:"cache_hit_seed"`
	CacheMissSeed      int64 `json:"cache_miss_seed"`
}

func (m *forYouMetrics) Snapshot() ForYouMetricsSnapshot {
	return ForYouMetricsSnapshot{
		CacheHitOwn:        m.CacheHitOwn.Load(),
		CacheMissOwn:       m.CacheMissOwn.Load(),
		CacheHitFollowing:  m.CacheHitFollowing.Load(),
		CacheMissFollowing: m.CacheMissFollowing.Load(),
		CacheHitCommunity:  m.CacheHitCommunity.Load(),
		CacheMissCommunity: m.CacheMissCommunity.Load(),
		CacheHitGlobal:     m.CacheHitGlobal.Load(),
		CacheMissGlobal:    m.CacheMissGlobal.Load(),
		CacheHitNewCount:   m.CacheHitNewCount.Load(),
		CacheMissNewCount:  m.CacheMissNewCount.Load(),
		CacheHitSeed:       m.CacheHitSeed.Load(),
		CacheMissSeed:      m.CacheMissSeed.Load(),
	}
}

// logSourceLatency emits one structured log line per source DB hit so a log
// aggregator can compute p50 / p95 latencies. Format is parseable by Loki /
// CloudWatch logs insights without extra config.
func logSourceLatency(source string, started time.Time, candidateCount int, err error) {
	ms := time.Since(started).Milliseconds()
	if err != nil {
		log.Printf("metric=feed_source_db_error source=%s duration_ms=%d candidates=%d err=%v", source, ms, candidateCount, err)
		return
	}
	log.Printf("metric=feed_source_db source=%s duration_ms=%d candidates=%d", source, ms, candidateCount)
}
