package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value interface{}
}

type lruCache struct {
	mu       sync.Mutex
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (cache *lruCache) Set(key Key, value interface{}) bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if item, ok := cache.items[key]; ok {
		item.Value = &cacheItem{key: key, value: value}
		cache.queue.MoveToFront(item)
		return true
	}

	cache.items[key] = cache.queue.PushFront(&cacheItem{key: key, value: value})
	if cache.queue.Len() > cache.capacity {
		last := cache.queue.Back()
		cache.queue.Remove(last)
		delete(cache.items, last.Value.(*cacheItem).key)
	}
	return false
}

func (cache *lruCache) Get(key Key) (interface{}, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	item, ok := cache.items[key]
	if !ok {
		return nil, false
	}
	cache.queue.MoveToFront(item)
	return item.Value.(*cacheItem).value, true
}

func (cache *lruCache) Clear() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.queue = NewList()
	cache.items = make(map[Key]*ListItem, cache.capacity)
}
