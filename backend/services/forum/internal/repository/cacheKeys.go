package repository

const (
	// HottestPostsCacheVersionKey is a monotonically increasing version used to invalidate
	// hot-post list caches (order=Hottest). Any change that may affect the hot ranking should
	// bump this version.
	HottestPostsCacheVersionKey = "forum:posts:hottest:ver"
)
