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

// globalThreadFeedsTag is registered on every cached thread feed key that has
// a global invalidation surface: thread:roots, thread:chain, thread:community,
// thread:all-communities, thread:my-communities, community:threads,
// event:threads. Used when a per-user mutation (profile rename, faculty
// change) affects fields embedded in cached threads across all viewers.
//
// Blast radius: drops every cached feed page. Same as the legacy
// DeleteByPatterns("thread:roots:*", "thread:community:*", ...) bundle.
func globalThreadFeedsTag() string {
	return "tag:feeds:global"
}

// viewerThreadFeedsTag covers every cached thread feed key keyed under a
// specific viewer (visibility depends on the viewer's follow graph or
// community membership). Used when a per-viewer mutation (follow/unfollow,
// community join/leave) invalidates that viewer's feeds without touching
// other viewers' cached feeds.
//
// Registered alongside the per-target tags for thread:userfeed/usermedia
// (which already use userFeedsTag(target)) so the requester-scoped pattern
// `thread:userfeed:*:req:<viewer>` is dropped by InvalidateTag.
func viewerThreadFeedsTag(viewerID uint) string {
	return fmt.Sprintf("tag:feeds:viewer:%d", viewerID)
}

// threadChainTag covers thread:chain:<rootID>:* — every cached page of a
// reply chain rooted at the given thread. Replaces the
// DeleteByPattern("thread:chain:<rootID>:*") SCAN.
func threadChainTag(rootID uint) string {
	return fmt.Sprintf("tag:thread:chain:%d", rootID)
}

// threadCommunityTag covers thread:community:<communityID>:* — every cached
// page of a community's thread list across viewers and sort modes. Replaces
// the DeleteByPattern("thread:community:<cid>:*") SCAN.
func threadCommunityTag(communityID uint) string {
	return fmt.Sprintf("tag:thread:community:%d", communityID)
}

// eventThreadsByEventTag covers event:threads:<eventID>:* — every cached
// page of one event's thread list across viewers, sort modes, filters.
// Replaces the DeleteByPattern("event:threads:<id>:*") SCAN.
func eventThreadsByEventTag(eventID string) string {
	return "tag:event:threads:event:" + eventID
}

// eventListsTag covers event:list:*, event:approved:list:*,
// event:pending:list:*. One tag for all three because they always invalidate
// together on event Save/Update/Delete.
func eventListsTag() string {
	return "tag:event:lists"
}

// Per-family global tags. Each replaces a high-frequency global SCAN pattern.
// Registered at the single Set site for the family alongside
// globalThreadFeedsTag so callers can choose precise (one family) or broad
// (all feed families) invalidation.
func threadRootsGlobalTag() string         { return "tag:thread:roots:global" }
func threadChainGlobalTag() string         { return "tag:thread:chain:global" }
func threadCommunityGlobalTag() string     { return "tag:thread:community:global" }
func threadAllCommunitiesGlobalTag() string { return "tag:thread:all-communities:global" }
func threadMyCommunitiesGlobalTag() string  { return "tag:thread:my-communities:global" }
func threadFollowingGlobalTag() string     { return "tag:thread:following:global" }
func communityThreadsGlobalTag() string    { return "tag:community:threads:global" }
func eventThreadsGlobalTag() string        { return "tag:event:threads:global" }

// followListUserTag covers follows:user:<uid>:<listType>:* for one user's
// following or followers list pages. listType is "following" or "followers".
// Replaces DeleteByPattern("follows:user:<uid>:<type>:*") in
// follow_repository.invalidateFollowLists.
func followListUserTag(userID uint, listType string) string {
	return fmt.Sprintf("tag:follows:user:%d:%s", userID, listType)
}

// interactionUserListTag covers <table>:user:<uid>:* for one user's bookmark/
// like/repost listings across pages. table is "likes" | "bookmarks" |
// "reposts". Replaces DeleteByPattern("<table>:user:<uid>:*") in
// interactionRepo.invalidateUser, event_repository.invalidateInteractionUserCaches,
// and user_profile_service.invalidateUserProfileCaches.
func interactionUserListTag(table string, userID uint) string {
	return fmt.Sprintf("tag:%s:user:%d", table, userID)
}

// communityMembersTag covers community:members:<cid>:*, community:users:<cid>:*,
// community:about:<cid>:*, community:about:topmembers:<cid>:* for one
// community. All four families invalidate together on member join/leave/role
// change.
func communityMembersTag(communityID uint) string {
	return fmt.Sprintf("tag:community:members:%d", communityID)
}

// communityMembersGlobalTag covers community user listings across ALL
// communities. Used by follow/unfollow which changes is_following on every
// community's user listing the followed user appears in. Coarser than
// enumerating affected communities; correctness == broader blast radius.
func communityMembersGlobalTag() string {
	return "tag:community:members:global"
}

// communityListsGlobalTag covers community:joined:* + community:created:*
// across all users. Used on community delete which invalidates every user's
// lists.
func communityListsGlobalTag() string {
	return "tag:community:lists:global"
}

// communityIDGlobalTag covers community:id:v2:* (every cached community
// detail). Used by user_profile_repository when avatar/profile changes affect
// embedded creator/member profiles snapshotted in cached community details.
func communityIDGlobalTag() string {
	return "tag:community:id:global"
}

// userFeedsGlobalTag covers thread:userfeed/usermedia/following:* across all
// users. Registered alongside userFeedsTag(uid) at every user-feed Set site.
// Used by faculty rename / user_settings / user_faculty global mutations.
func userFeedsGlobalTag() string {
	return "tag:user:feeds:global"
}

// userForYouGlobalTag covers feed:foryou:user:*:* across all users. Used
// when a global mutation invalidates every viewer's foryou candidates
// (faculty rename, event update, etc).
func userForYouGlobalTag() string {
	return "tag:user:foryou:global"
}

// threadIDGlobalTag covers thread:id:<n> across all threads. Registered at
// every Set of the per-thread cache. Used when a global mutation affects
// fields embedded in cached threads (faculty rename, community update).
func threadIDGlobalTag() string {
	return "tag:thread:id:global"
}

// profileUserGlobalTag covers profile:user:<uid>:* across all users.
func profileUserGlobalTag() string {
	return "tag:profile:user:global"
}

// userFacultyGlobalTag covers userfaculty:user:<uid>:* across all users.
func userFacultyGlobalTag() string {
	return "tag:userfaculty:user:global"
}

// userMetaGlobalTag covers user:meta:<uid> across all users.
func userMetaGlobalTag() string {
	return "tag:user:meta:global"
}

// notificationSenderGlobalTag covers notification:sender:<uid> across all users.
func notificationSenderGlobalTag() string {
	return "tag:notification:sender:global"
}

// pollUserVotesTag covers poll:<pollID>:user:*. Registered at every
// pollUserVoteKey Set. Replaces DeleteByPattern in ThreadPollRepository.Delete.
func pollUserVotesTag(pollID uint) string {
	return fmt.Sprintf("tag:poll:%d:user-votes", pollID)
}

// facultyListTag covers faculty:list:page:*:limit:* (paginated faculty lists).
// Registered at FacultyRepository.FindAll Set site.
func facultyListTag() string {
	return "tag:faculty:list"
}

// trendingHashtagsTag covers trending_hashtags:* (the trending list cache
// family). Single tag for the family.
func trendingHashtagsTag() string {
	return "tag:trending-hashtags"
}

// reportsIDTag covers reports:id:* (paginated reports list cache).
func reportsIDTag() string {
	return "tag:reports:id"
}

// communityRequestsGlobalTag covers community:requests:* across all communities.
// Used by user_profile_service / premium expiry / order approval for blanket
// invalidation when enumerating affected communities is impractical.
func communityRequestsGlobalTag() string {
	return "tag:community:requests:global"
}

// exploreGlobalTag covers every explore:* cache key (search, top, latest,
// people, communities, suggested_people). One tag for the family.
func exploreGlobalTag() string {
	return "tag:explore:global"
}

// exploreTopTag covers explore:top:* (top posts section only).
func exploreTopTag() string {
	return "tag:explore:top"
}

// exploreLatestTag covers explore:latest:* (latest posts section only).
func exploreLatestTag() string {
	return "tag:explore:latest"
}

// communityListsByUserTag covers community:created:<uid>:* and
// community:joined:<uid>:* for one user's community membership listings.
func communityListsByUserTag(userID uint) string {
	return fmt.Sprintf("tag:community:lists:user:%d", userID)
}

// communityRequestsTag covers community:requests:<cid>:* for one community's
// pending-join-request list.
func communityRequestsTag(communityID uint) string {
	return fmt.Sprintf("tag:community:requests:%d", communityID)
}

// communityPinsUserTag covers community:pins:user:<uid>:* for one user's
// pinned-communities list.
func communityPinsUserTag(userID uint) string {
	return fmt.Sprintf("tag:community:pins:user:%d", userID)
}

// eventJoinedUserTag covers event:joined:user:<uid>:* for one user's
// joined-events list.
func eventJoinedUserTag(userID uint) string {
	return fmt.Sprintf("tag:event:joined:user:%d", userID)
}

// eventPinsUserTag covers event:pins:user:<uid>:* for one user's pinned-events list.
func eventPinsUserTag(userID uint) string {
	return fmt.Sprintf("tag:event:pins:user:%d", userID)
}
