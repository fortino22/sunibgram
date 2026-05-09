package repository

import (
	"context"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

type threadCached struct {
	db        *threadDB
	cache     *cache.Cache
	ttl       time.Duration
	likes     *interactionRepo
	bookmarks *interactionRepo
	reposts   *interactionRepo
}

type threadDB struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

const threadFeedTTL = 60 * time.Second

func preloadThreadUser(db *gorm.DB) *gorm.DB {
	return db.Select("user_id", "name", "user_name", "is_premium")
}

func (db *threadDB) attachFaculties(threads []model.Thread) {
	userIDs := make(map[uint]struct{})
	for i := range threads {
		userIDs[threads[i].UserID] = struct{}{}
		if threads[i].OriginalThread != nil {
			userIDs[threads[i].OriginalThread.UserID] = struct{}{}
		}
	}
	if len(userIDs) == 0 {
		return
	}
	ids := make([]uint, 0, len(userIDs))
	for id := range userIDs {
		ids = append(ids, id)
	}
	var faculties []model.UserFaculty
	db.readDB.Preload("Faculty").Where("user_id IN ?", ids).Find(&faculties)
	fm := make(map[uint]*model.UserFaculty, len(faculties))
	for i := range faculties {
		fm[faculties[i].UserID] = &faculties[i]
	}
	for i := range threads {
		if uf, ok := fm[threads[i].UserID]; ok {
			threads[i].User.UserFaculty = uf
		}
		if threads[i].OriginalThread != nil {
			if uf, ok := fm[threads[i].OriginalThread.UserID]; ok {
				threads[i].OriginalThread.User.UserFaculty = uf
			}
		}
	}
}

func NewThreadRepository(writeDB, readDB *gorm.DB, c *cache.Cache) (domainrepo.IThreadRepository, func(uint, string, int)) {
	tdb := &threadDB{writeDB: writeDB, readDB: readDB}
	r := &threadCached{
		db:        tdb,
		cache:     c,
		ttl:       cache.DefaultTTL,
		likes:     newInteractionRepo(writeDB, readDB, c, "likes", tdb),
		bookmarks: newInteractionRepo(writeDB, readDB, c, "bookmarks", tdb),
		reposts:   newInteractionRepo(writeDB, readDB, c, "reposts", tdb),
	}
	return r, r.patchCachedCount
}

func (r *threadCached) Save(thread *model.Thread) error {
	if err := r.db.writeDB.Create(thread).Error; err != nil {
		return err
	}
	ctx := context.Background()
	if thread.Type == model.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("repost_count + 1"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", 1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
		var orig model.Thread
		if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, *thread.OriginalThreadID).Error; err == nil {
			r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
		}
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.ParentID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
		r.patchCachedCount(*thread.ParentID, "comment_count", 1)
	}
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	invalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *threadCached) InvalidateThread(threadID uint) {
	ctx := context.Background()
	thread, err := r.FindByID(threadID)
	if err != nil {
		r.cache.Delete(ctx, threadKey(threadID))
		return
	}
	r.invalidate(ctx, threadID, thread.RootID, thread.CommunityID)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	r.invalidateInteractions(ctx, threadID)
}

func (r *threadCached) FindByID(id uint) (*model.Thread, error) {
	ctx := context.Background()
	var thread model.Thread
	key := threadKey(id)
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").First(&thread, id).Error; err != nil {
			return nil, err
		}
		threads := []model.Thread{thread}
		r.db.attachFaculties(threads)
		thread = threads[0]
		r.cache.Set(ctx, key, thread, r.ttl)
	} else {
		threads := []model.Thread{thread}
		r.db.attachFaculties(threads)
		thread = threads[0]
	}
	return &thread, nil
}

func (r *threadCached) FindRoots(page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadRootsKey(page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.type IN ?", []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		r.db.attachFaculties(threads)
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadChainKey(rootID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").
			Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
			Where("threads.thread_id = ? OR threads.root_id = ?", rootID, rootID).
			Where("threads.community_id IS NULL OR communities.visibility = ?", model.CommunityVisibilityPublic).
			Order("threads.created_at ASC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		r.db.attachFaculties(threads)
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}

func (r *threadCached) findUserFeed(userID uint, types []string, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	var total int64
	r.db.readDB.Model(&model.Thread{}).
		Where("user_id = ? AND type IN ?", userID, types).
		Count(&total)

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").
		Where("user_id = ? AND type IN ?", userID, types).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}
	r.db.attachFaculties(threads)

	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, 0, err
		}
		r.db.attachFaculties(parentThreads)
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	parents := make([]*model.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentMap[*t.ParentID]
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindPostsByUserID(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findUserFeed(userID, []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeRepost),
	}, page, limit)
}

func (r *threadCached) FindReplyTabByUserID(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	return r.findUserFeed(userID, []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeReply),
		string(model.ThreadTypeRepost),
	}, page, limit)
}

func (r *threadCached) ExistsRepost(userID, originalThreadID uint) (bool, error) {
	var count int64
	err := r.db.readDB.Model(&model.Thread{}).
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, model.ThreadTypeRepost, originalThreadID).
		Count(&count).Error
	return count > 0, err
}

func (r *threadCached) DeleteRepost(userID, originalThreadID uint) error {
	var repost model.Thread
	if err := r.db.readDB.
		Where("user_id = ? AND type = ? AND original_thread_id = ?", userID, model.ThreadTypeRepost, originalThreadID).
		First(&repost).Error; err != nil {
		return err
	}

	if err := r.db.writeDB.Delete(&repost).Error; err != nil {
		return err
	}

	r.db.writeDB.Model(&model.Thread{}).
		Where("thread_id = ?", originalThreadID).
		UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
	r.patchCachedCount(originalThreadID, "repost_count", -1)

	ctx := context.Background()
	r.invalidate(ctx, repost.ThreadID, nil, nil)
	r.invalidateInteractions(ctx, originalThreadID)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	var orig model.Thread
	if err := r.db.readDB.Select("thread_id, root_id, community_id").First(&orig, originalThreadID).Error; err == nil {
		r.invalidate(ctx, orig.ThreadID, orig.RootID, orig.CommunityID)
	}
	invalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *threadCached) Update(thread *model.Thread) error {
	if err := r.db.writeDB.Model(thread).
		UpdateColumns(map[string]any{
			"content":    thread.Content,
			"updated_at": gorm.Expr("NOW()"),
		}).Error; err != nil {
		return err
	}
	ctx := context.Background()
	r.cache.Delete(ctx, threadKey(thread.ThreadID))
	r.invalidate(ctx, thread.ThreadID, thread.RootID, thread.CommunityID)
	r.invalidateInteractions(ctx, thread.ThreadID)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")
	invalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *threadCached) patchCachedCount(threadID uint, column string, delta int) {
	ctx := context.Background()
	key := threadKey(threadID)
	var thread model.Thread
	if err := r.cache.Get(ctx, key, &thread); err != nil {
		if err := r.db.readDB.Preload("User", preloadThreadUser).First(&thread, threadID).Error; err != nil {
			return
		}
	}
	switch column {
	case "like_count":
		thread.LikeCount = max(thread.LikeCount+delta, 0)
	case "repost_count":
		thread.RepostCount = max(thread.RepostCount+delta, 0)
	case "bookmark_count":
		thread.BookmarkCount = max(thread.BookmarkCount+delta, 0)
	case "comment_count":
		thread.CommentCount = max(thread.CommentCount+delta, 0)
	}
	r.cache.Set(ctx, key, thread, r.ttl)
}

func (r *threadCached) Delete(id uint) error {
	var thread model.Thread
	r.db.readDB.Select("thread_id, parent_id, root_id, type, original_thread_id, community_id").First(&thread, id)

	ctx := context.Background()
	if thread.Type == model.ThreadTypeRepost && thread.OriginalThreadID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.OriginalThreadID).
			UpdateColumn("repost_count", gorm.Expr("GREATEST(repost_count - 1, 0)"))
		r.patchCachedCount(*thread.OriginalThreadID, "repost_count", -1)
		r.invalidateInteractions(ctx, *thread.OriginalThreadID)
	}
	if thread.ParentID != nil {
		r.db.writeDB.Model(&model.Thread{}).
			Where("thread_id = ?", *thread.ParentID).
			UpdateColumn("comment_count", gorm.Expr("GREATEST(comment_count - 1, 0)"))
		r.patchCachedCount(*thread.ParentID, "comment_count", -1)
	}

	r.invalidate(ctx, id, thread.RootID, thread.CommunityID)
	r.invalidateInteractions(ctx, id)
	r.cache.DeleteByPattern(ctx, "thread:roots:*")

	r.likes.deleteByThreadID(id)
	r.bookmarks.deleteByThreadID(id)
	r.reposts.deleteByThreadID(id)

	if err := r.db.writeDB.Delete(&model.Thread{}, id).Error; err != nil {
		return err
	}
	invalidateExploreCache(ctx, r.cache)
	return nil
}

func (r *threadCached) FindFollowingFeed(userID uint, page, limit int) ([]model.Thread, []*model.Thread, int64, error) {
	types := []string{
		string(model.ThreadTypePost),
		string(model.ThreadTypeRepost),
		string(model.ThreadTypeReply),
	}

	var total int64
	r.db.readDB.Model(&model.Thread{}).
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Count(&total)

	offset := (page - 1) * limit
	var threads []model.Thread
	if err := r.db.readDB.
		Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).Preload("Community").Preload("OriginalThread.Community").
		Joins("JOIN user_relationships ON user_relationships.following_id = threads.user_id").
		Joins("LEFT JOIN communities ON communities.community_id = threads.community_id").
		Joins("LEFT JOIN community_members ON community_members.community_id = threads.community_id AND community_members.user_id = ?", userID).
		Where("user_relationships.follower_id = ? AND threads.type IN ?", userID, types).
		Where("threads.community_id IS NULL OR communities.visibility = ? OR community_members.user_id = ?",
			model.CommunityVisibilityPublic, userID).
		Order("threads.created_at DESC").
		Offset(offset).Limit(limit).
		Find(&threads).Error; err != nil {
		return nil, nil, 0, err
	}
	r.db.attachFaculties(threads)

	parentIDSet := make(map[uint]struct{})
	for _, t := range threads {
		if t.ParentID != nil {
			parentIDSet[*t.ParentID] = struct{}{}
		}
	}
	parentMap := make(map[uint]*model.Thread)
	if len(parentIDSet) > 0 {
		pids := make([]uint, 0, len(parentIDSet))
		for id := range parentIDSet {
			pids = append(pids, id)
		}
		var parentThreads []model.Thread
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("Community").Where("thread_id IN ?", pids).Find(&parentThreads).Error; err != nil {
			return nil, nil, 0, err
		}
		r.db.attachFaculties(parentThreads)
		for i := range parentThreads {
			parentMap[parentThreads[i].ThreadID] = &parentThreads[i]
		}
	}

	parents := make([]*model.Thread, len(threads))
	for i, t := range threads {
		if t.ParentID != nil {
			parents[i] = parentMap[*t.ParentID]
		}
	}
	return threads, parents, total, nil
}

func (r *threadCached) FindByCommunityID(communityID uint, page, limit int) ([]model.Thread, int64, error) {
	type cached struct {
		Threads []model.Thread
		Total   int64
	}
	ctx := context.Background()
	var result cached
	key := threadCommunityKey(communityID, page, limit)
	if err := r.cache.Get(ctx, key, &result); err != nil {
		var threads []model.Thread
		var total int64
		r.db.readDB.Model(&model.Thread{}).
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Count(&total)
		offset := (page - 1) * limit
		if err := r.db.readDB.Preload("User", preloadThreadUser).Preload("OriginalThread.User", preloadThreadUser).
			Where("community_id = ? AND type IN ?", communityID, []string{string(model.ThreadTypePost), string(model.ThreadTypeRepost)}).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&threads).Error; err != nil {
			return nil, 0, err
		}
		r.db.attachFaculties(threads)
		result = cached{Threads: threads, Total: total}
		r.cache.Set(ctx, key, result, threadFeedTTL)
	}
	return result.Threads, result.Total, nil
}
