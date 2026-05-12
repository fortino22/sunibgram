package helper

func NormalizePageLimit(page, limit, defaultLimit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = defaultLimit
	}
	return page, limit
}
