package cache

import (
	"container/list"
	"sync"
	"time"
)

type LRU[K comparable, V any] struct {
	mu         sync.Mutex
	maxEntries int
	ll         *list.List
	cache      map[K]*list.Element
}

type lruEntry[K comparable, V any] struct {
	key      K
	value    V
	expireAt time.Time
}

func NewLRU[K comparable, V any](maxEntries int) *LRU[K, V] {
	if maxEntries <= 0 {
		maxEntries = 1024
	}
	return &LRU[K, V]{
		maxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[K]*list.Element, maxEntries),
	}
}

func (l *LRU[K, V]) Get(key K) (V, bool) {
	var zero V
	if l == nil {
		return zero, false
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if ele, ok := l.cache[key]; ok {
		ent := ele.Value.(*lruEntry[K, V])
		if !ent.expireAt.IsZero() && time.Now().After(ent.expireAt) {
			l.ll.Remove(ele)
			delete(l.cache, key)
			return zero, false
		}
		l.ll.MoveToFront(ele)
		return ent.value, true
	}
	return zero, false
}

func (l *LRU[K, V]) Add(key K, value V, ttl time.Duration) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	if ele, ok := l.cache[key]; ok {
		ent := ele.Value.(*lruEntry[K, V])
		ent.value = value
		if ttl > 0 {
			ent.expireAt = time.Now().Add(ttl)
		} else {
			ent.expireAt = time.Time{}
		}
		l.ll.MoveToFront(ele)
		return
	}

	ent := &lruEntry[K, V]{key: key, value: value}
	if ttl > 0 {
		ent.expireAt = time.Now().Add(ttl)
	}
	ele := l.ll.PushFront(ent)
	l.cache[key] = ele

	if l.ll.Len() > l.maxEntries {
		l.removeOldest()
	}
}

func (l *LRU[K, V]) Remove(key K) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if ele, ok := l.cache[key]; ok {
		l.ll.Remove(ele)
		delete(l.cache, key)
	}
}

func (l *LRU[K, V]) Purge() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ll.Init()
	clear(l.cache)
}

func (l *LRU[K, V]) removeOldest() {
	ele := l.ll.Back()
	if ele == nil {
		return
	}
	l.ll.Remove(ele)
	ent := ele.Value.(*lruEntry[K, V])
	delete(l.cache, ent.key)
}
