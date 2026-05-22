package repository

import "fmt"

// Tag key builders for reverse-index invalidation (cache phase 2).
//
// Convention: `tag:<entity>:<id>[:scope]`. Tag namespace is prefixed `tag:`
// so it never collides with cached data keys. Scope segment is optional and
// used when one entity has multiple disjoint invalidation surfaces (e.g.
// a user's feed pages vs the same user's foryou source caches).
//
// Add new builders here. Never inline tag-string construction at call sites.

// userFeedsTag covers all cached feeds keyed under userID's profile view:
// thread:userfeed:<uid>:*, thread:usermedia:<uid>:*, thread:following:<uid>:*.
func userFeedsTag(userID uint) string {
	return fmt.Sprintf("tag:user:%d:feeds", userID)
}

// userForYouTag covers foryou source + newcount caches for a user.
// Does NOT cover the foryou seed (deterministic key, deleted directly).
func userForYouTag(userID uint) string {
	return fmt.Sprintf("tag:user:%d:foryou", userID)
}

// threadFeedsTag covers feed entries that include a given thread. Replaces
// the legacy `thread:feed_keys:<id>` SET naming.
func threadFeedsTag(threadID uint) string {
	return fmt.Sprintf("tag:thread:%d:feeds", threadID)
}
