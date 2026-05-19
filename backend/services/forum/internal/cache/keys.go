package cache

import (
	"fmt"
	"time"
)

func postVisitKey(postID int64) string {
	return fmt.Sprintf("post:visit:%d", postID)
}

func postDetailKey(postID int64) string {
	return fmt.Sprintf("post:%d", postID)
}

func postLockKey(postID int64) string {
	return fmt.Sprintf("lock:post:%d", postID)
}

func postHotBucketKey(t time.Time) string {
	return fmt.Sprintf("post:hot:%s", t.Format("2006010215"))
}

func postHotAggregateKey(topN int64) string {
	return fmt.Sprintf("post:hot:24h:top:%d", topN)
}

func commentListKey(postID int64) string {
	return fmt.Sprintf("comments:post:%d", postID)
}

func commentDetailKey(commentID int64) string {
	return fmt.Sprintf("comment:%d", commentID)
}

func commentListLockKey(postID int64) string {
	return fmt.Sprintf("lock:comments:post:%d", postID)
}
