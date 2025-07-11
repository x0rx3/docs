package service

import (
	"sync"
	"time"
)

type item struct {
	value        any
	liveDuration time.Time
	tags         []string
}

type InternalCache struct {
	sync.RWMutex
	items     map[string]item
	tagsIndex map[string]map[string]struct{}
}

func NewInternalCache() *InternalCache {
	return &InternalCache{
		items:     make(map[string]item),
		tagsIndex: make(map[string]map[string]struct{}),
	}
}

func (inst *InternalCache) Get(key string) (any, bool) {
	inst.Lock()
	defer inst.Unlock()

	item, found := inst.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(item.liveDuration) {
		go inst.deleteKey(key)
		return nil, false
	}

	return item.value, true
}

func (inst *InternalCache) Put(k string, value any, ttl time.Duration, tags []string) {
	inst.Lock()
	defer inst.Unlock()

	if item, exists := inst.items[k]; exists {
		inst.deleteItemFromTag(k, item.tags)
	}

	inst.items[k] = item{
		value:        value,
		liveDuration: time.Now().Add(ttl),
		tags:         tags,
	}

	for _, tg := range tags {
		if _, exists := inst.tagsIndex[tg]; !exists {
			inst.tagsIndex[tg] = make(map[string]struct{})
		}
		inst.tagsIndex[tg][k] = struct{}{}
	}
}

func (inst *InternalCache) Invalidate(key string) {
	inst.Lock()
	defer inst.Unlock()
	inst.deleteKey(key)
}

func (inst *InternalCache) InvalidateByTag(tag string) {
	inst.Lock()
	defer inst.Unlock()

	if keys, exists := inst.tagsIndex[tag]; exists {
		for key := range keys {
			inst.deleteKey(key)
		}
		delete(inst.tagsIndex, tag)
	}
}

func (inst *InternalCache) InvalidateByTags(tags []string) {
	inst.Lock()
	defer inst.Unlock()

	for _, tag := range tags {
		if keys, exists := inst.tagsIndex[tag]; exists {
			for key := range keys {
				inst.deleteKey(key)
			}
			delete(inst.tagsIndex, tag)
		}
	}

}

func (inst *InternalCache) CleanExpired() {
	inst.Lock()
	defer inst.Unlock()

	now := time.Now()
	for key, item := range inst.items {
		if now.After(item.liveDuration) {
			inst.deleteKey(key)
		}
	}
}

func (inst *InternalCache) deleteKey(key string) {
	if item, exists := inst.items[key]; exists {
		inst.deleteItemFromTag(key, item.tags)
		delete(inst.items, key)
	}
}

func (inst *InternalCache) deleteItemFromTag(key string, tags []string) {
	for _, tag := range tags {
		if keys, exists := inst.tagsIndex[tag]; exists {
			delete(keys, key)
			if len(keys) == 0 {
				delete(inst.tagsIndex, tag)
			}
		}
	}
}
