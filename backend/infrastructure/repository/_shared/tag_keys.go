package infrarepoShared

import "fmt"

// Tag key builders for reverse-index invalidation (cache phase 2).
//
// Convention: `tag:<entity>:<id>[:scope]`. Tag namespace is prefixed `tag:`
// so it never collides with cached data keys. Scope segment is optional and
// used when one entity has multiple disjoint invalidation surfaces (e.g.
// a user's feed pages vs the same user's foryou source caches).
//
// Add new builders here. Never inline tag-string construction at call sites.

// UserFeedsTag covers all cached feeds keyed under userID's profile view:
// thread:userfeed:<uid>:*, thread:usermedia:<uid>:*, thread:following:<uid>:*.
func UserFeedsTag(userID uint) string {
	return fmt.Sprintf("tag:user:%d:feeds", userID)
}

// UserForYouTag covers foryou source + newcount caches for a user.
// Does NOT cover the foryou seed (deterministic key, deleted directly).
func UserForYouTag(userID uint) string {
	return fmt.Sprintf("tag:user:%d:foryou", userID)
}

// ThreadFeedsTag covers feed entries that include a given thread. Replaces
// the legacy `thread:feed_keys:<id>` SET naming.
func ThreadFeedsTag(threadID uint) string {
	return fmt.Sprintf("tag:thread:%d:feeds", threadID)
}

// GlobalThreadFeedsTag is registered on every cached thread feed key that has
// a global invalidation surface: thread:roots, thread:chain, thread:community,
// thread:all-communities, thread:my-communities, community:threads,
// event:threads. Used when a per-user mutation (profile rename, faculty
// change) affects fields embedded in cached threads across all viewers.
//
// Blast radius: drops every cached feed page. Same as the legacy
// DeleteByPatterns("thread:roots:*", "thread:community:*", ...) bundle.
func GlobalThreadFeedsTag() string {
	return "tag:feeds:global"
}

// ViewerThreadFeedsTag covers every cached thread feed key keyed under a
// specific viewer (visibility depends on the viewer's follow graph or
// community membership). Used when a per-viewer mutation (follow/unfollow,
// community join/leave) invalidates that viewer's feeds without touching
// other viewers' cached feeds.
//
// Registered alongside the per-target tags for thread:userfeed/usermedia
// (which already use UserFeedsTag(target)) so the requester-scoped pattern
// `thread:userfeed:*:req:<viewer>` is dropped by InvalidateTag.
func ViewerThreadFeedsTag(viewerID uint) string {
	return fmt.Sprintf("tag:feeds:viewer:%d", viewerID)
}

// ThreadChainTag covers thread:chain:<rootID>:* — every cached page of a
// reply chain rooted at the given thread. Replaces the
// DeleteByPattern("thread:chain:<rootID>:*") SCAN.
func ThreadChainTag(rootID uint) string {
	return fmt.Sprintf("tag:thread:chain:%d", rootID)
}

// ThreadCommunityTag covers thread:community:<communityID>:* — every cached
// page of a community's thread list across viewers and sort modes. Replaces
// the DeleteByPattern("thread:community:<cid>:*") SCAN.
func ThreadCommunityTag(communityID uint) string {
	return fmt.Sprintf("tag:thread:community:%d", communityID)
}

// EventThreadsByEventTag covers event:threads:<eventID>:* — every cached
// page of one event's thread list across viewers, sort modes, filters.
// Replaces the DeleteByPattern("event:threads:<id>:*") SCAN.
func EventThreadsByEventTag(eventID string) string {
	return "tag:event:threads:event:" + eventID
}

// EventListsTag covers event:list:*, event:approved:list:*,
// event:pending:list:*. One tag for all three because they always invalidate
// together on event Save/Update/Delete.
func EventListsTag() string {
	return "tag:event:lists"
}

// Per-family global tags. Each replaces a high-frequency global SCAN pattern.
// Registered at the single Set site for the family alongside
// GlobalThreadFeedsTag so callers can choose precise (one family) or broad
// (all feed families) invalidation.
func ThreadRootsGlobalTag() string          { return "tag:thread:roots:global" }
func ThreadChainGlobalTag() string          { return "tag:thread:chain:global" }
func ThreadCommunityGlobalTag() string      { return "tag:thread:community:global" }
func ThreadAllCommunitiesGlobalTag() string { return "tag:thread:all-communities:global" }
func ThreadMyCommunitiesGlobalTag() string  { return "tag:thread:my-communities:global" }
func ThreadFollowingGlobalTag() string      { return "tag:thread:following:global" }
func CommunityThreadsGlobalTag() string     { return "tag:community:threads:global" }
func EventThreadsGlobalTag() string         { return "tag:event:threads:global" }

// FollowListUserTag covers follows:user:<uid>:<listType>:* for one user's
// following or followers list pages. listType is "following" or "followers".
// Replaces DeleteByPattern("follows:user:<uid>:<type>:*") in
// follow_repository.invalidateFollowLists.
func FollowListUserTag(userID uint, listType string) string {
	return fmt.Sprintf("tag:follows:user:%d:%s", userID, listType)
}

// FollowListGlobalTag covers every cached follow-list page across all users.
// Registered alongside FollowListUserTag(uid, type) at each Set site.
// Replaces DeleteByPattern("follows:user:*") in user_repository.invalidateUserPremiumCaches.
func FollowListGlobalTag() string {
	return "tag:follows:user:global"
}

// InteractionUserListTag covers <table>:user:<uid>:* for one user's bookmark/
// like/repost listings across pages. table is "likes" | "bookmarks" |
// "reposts". Replaces DeleteByPattern("<table>:user:<uid>:*") in
// InteractionRepo.InvalidateUser, event_repository.invalidateInteractionUserCaches,
// and user_profile_service.invalidateUserProfileCaches.
func InteractionUserListTag(table string, userID uint) string {
	return fmt.Sprintf("tag:%s:user:%d", table, userID)
}

// CommunityMembersTag covers community:members:<cid>:*, community:users:<cid>:*,
// community:about:<cid>:*, community:about:topmembers:<cid>:* for one
// community. All four families invalidate together on member join/leave/role
// change.
func CommunityMembersTag(communityID uint) string {
	return fmt.Sprintf("tag:community:members:%d", communityID)
}

// CommunityMembersGlobalTag covers community user listings across ALL
// communities. Used by follow/unfollow which changes is_following on every
// community's user listing the followed user appears in. Coarser than
// enumerating affected communities; correctness == broader blast radius.
func CommunityMembersGlobalTag() string {
	return "tag:community:members:global"
}

// CommunityListsGlobalTag covers community:joined:* + community:created:*
// across all users. Used on community delete which invalidates every user's
// lists.
func CommunityListsGlobalTag() string {
	return "tag:community:lists:global"
}

// CommunityIDGlobalTag covers community:id:v2:* (every cached community
// detail). Used by user_profile_repository when avatar/profile changes affect
// embedded creator/member profiles snapshotted in cached community details.
func CommunityIDGlobalTag() string {
	return "tag:community:id:global"
}

// userFeedsGlobalTag covers thread:userfeed/usermedia/following:* across all
// users. Registered alongside UserFeedsTag(uid) at every user-feed Set site.
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

// NotificationSenderGlobalTag covers notification:sender:<uid> across all users.
func NotificationSenderGlobalTag() string {
	return "tag:notification:sender:global"
}

// PollUserVotesTag covers poll:<pollID>:user:*. Registered at every
// threadpollrepo.PollUserVoteKey Set. Replaces DeleteByPattern in ThreadPollRepository.Delete.
func PollUserVotesTag(pollID uint) string {
	return fmt.Sprintf("tag:poll:%d:user-votes", pollID)
}

// FacultyListTag covers faculty:list:page:*:limit:* (paginated faculty lists).
// Registered at FacultyRepository.FindAll Set site.
func FacultyListTag() string {
	return "tag:faculty:list"
}

// TrendingHashtagsTag covers trending_hashtags:* (the trending list cache
// family). Single tag for the family.
func TrendingHashtagsTag() string {
	return "tag:trending-hashtags"
}

// ReportsIDTag covers reports:id:* (paginated reports list cache).
func ReportsIDTag() string {
	return "tag:reports:id"
}

// CommunityRequestsGlobalTag covers community:requests:* across all communities.
// Used by user_profile_service / premium expiry / order approval for blanket
// invalidation when enumerating affected communities is impractical.
func CommunityRequestsGlobalTag() string {
	return "tag:community:requests:global"
}

// ExploreGlobalTag covers every explore:* cache key (search, top, latest,
// people, communities, suggested_people). One tag for the family.
func ExploreGlobalTag() string {
	return "tag:explore:global"
}

// ExploreTopTag covers explore:top:* (top posts section only).
func ExploreTopTag() string {
	return "tag:explore:top"
}

// ExploreLatestTag covers explore:latest:* (latest posts section only).
func ExploreLatestTag() string {
	return "tag:explore:latest"
}

// CommunityListsByUserTag covers community:created:<uid>:* and
// community:joined:<uid>:* for one user's community membership listings.
func CommunityListsByUserTag(userID uint) string {
	return fmt.Sprintf("tag:community:lists:user:%d", userID)
}

// CommunityRequestsTag covers community:requests:<cid>:* for one community's
// pending-join-request list.
func CommunityRequestsTag(communityID uint) string {
	return fmt.Sprintf("tag:community:requests:%d", communityID)
}

// CommunityPinsUserTag covers community:pins:user:<uid>:* for one user's
// pinned-communities list.
func CommunityPinsUserTag(userID uint) string {
	return fmt.Sprintf("tag:community:pins:user:%d", userID)
}

// EventJoinedUserTag covers event:joined:user:<uid>:* for one user's
// joined-events list.
func EventJoinedUserTag(userID uint) string {
	return fmt.Sprintf("tag:event:joined:user:%d", userID)
}

// EventPinsUserTag covers event:pins:user:<uid>:* for one user's pinned-events list.
func EventPinsUserTag(userID uint) string {
	return fmt.Sprintf("tag:event:pins:user:%d", userID)
}
