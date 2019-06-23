package base

import (
	"container/list"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

// A SizedEntry is an entry within the LRUCache that knows its own size.
type SizedEntry interface {
	// Release will be called upon the entry being evicted from the cache.
	Release()

	// Size returns the number of bytes consumed by the entry.
	Size() Byte
}

// A SizedEntryRef is a wrapper around a SizedEntry.
type SizedEntryRef struct {
	Value      SizedEntry
	lruCache   *LRUCache
	cacheEntry *lruCacheEntry
}

type SizedEntryFactory func(key string) (SizedEntry, error)

// An lruCacheEntry is an entry into an LRUCache.
//
// refCount is zero if and only if listElement is non-nil.
type lruCacheEntry struct {
	refCount    int32
	listElement *list.Element
	sizedEntry  SizedEntry
	key         string
}

// LRUCache handles a pool of sized resources. It has a fixed maximum size with
// a least-recently used eviction policy.
type LRUCache struct {
	sync.Mutex
	mapping       map[string]*lruCacheEntry
	evictList     *list.List
	totalSize     Byte
	evictableSize Byte
	sizeLimit     Byte
}

// NewLRUCache returns an empty LRUCache with the provided size limit.
func NewLRUCache(sizeLimit Byte) *LRUCache {
	return &LRUCache{
		mapping:   make(map[string]*lruCacheEntry),
		evictList: list.New(),
		sizeLimit: sizeLimit,
	}
}

func (c *LRUCache) evictLocked() {
	for c.evictList.Len() > 0 && c.totalSize.Bytes() > c.sizeLimit.Bytes() {
		element := c.evictList.Back()
		cacheEntry := element.Value.(*lruCacheEntry)

		if cacheEntry.refCount != 0 {
			panic(errors.Errorf("Invalid refcount for LRU cache entry: %d", cacheEntry.refCount))
		}
		if cacheEntry.listElement != element {
			panic(errors.Errorf(
				"Invalid refcount for LRU cache list element: %p != %p",
				cacheEntry.listElement,
				element,
			))
		}

		c.totalSize = Byte(c.totalSize.Bytes() - cacheEntry.sizedEntry.Size().Bytes())
		c.evictableSize = Byte(c.evictableSize.Bytes() - cacheEntry.sizedEntry.Size().Bytes())

		cacheEntry.listElement = nil
		c.evictList.Remove(element)
		delete(c.mapping, cacheEntry.key)

		cacheEntry.sizedEntry.Release()
	}
}

func (c *LRUCache) reserveLocked(size Byte) {
	c.totalSize = Byte(c.totalSize.Bytes() + size.Bytes())
	c.evictLocked()
}

// Get atomically gets a previously-created entry if it was found in the cache,
// or a newly-created one otherwise. It is the caller's responsibility to call
// Put() with the returned SizedEntryRef method once it's no longer needed so
// that the underlying resource can be evicted from the cache, if needed.
func (c *LRUCache) Get(
	key string,
	factory SizedEntryFactory,
) (*SizedEntryRef, error) {
	c.Lock()
	defer c.Unlock()

	if cacheEntry, ok := c.mapping[key]; ok {
		if atomic.AddInt32(&cacheEntry.refCount, 1) == 1 {
			if cacheEntry.listElement == nil {
				panic(errors.New("Invalid nil LRU cache list element"))
			}
			c.evictList.Remove(cacheEntry.listElement)
			c.evictableSize = Byte(c.evictableSize.Bytes() - cacheEntry.sizedEntry.Size().Bytes())
			cacheEntry.listElement = nil
		}
		return &SizedEntryRef{
			Value:      cacheEntry.sizedEntry,
			lruCache:   c,
			cacheEntry: cacheEntry,
		}, nil
	}

	value, err := factory(key)
	if err != nil {
		return nil, err
	}

	c.reserveLocked(value.Size())
	cacheEntry := &lruCacheEntry{
		refCount:   1,
		sizedEntry: value,
		key:        key,
	}

	c.mapping[key] = cacheEntry

	return &SizedEntryRef{
		Value:      value,
		lruCache:   c,
		cacheEntry: cacheEntry,
	}, nil
}

// Put marks a SizedEntryRef as no longer being referred to, so that it can be
// considered for eviction.
func (c *LRUCache) Put(r *SizedEntryRef) {
	c.Lock()
	defer c.Unlock()

	if atomic.AddInt32(&r.cacheEntry.refCount, -1) != 0 {
		return
	}

	if r.cacheEntry.listElement != nil {
		panic(errors.Errorf(
			"Invalid non-nil LRU cache list element: %p",
			r.cacheEntry.listElement,
		))
	}

	r.cacheEntry.listElement = c.evictList.PushFront(r.cacheEntry)
	c.evictableSize = Byte(c.evictableSize.Bytes() + r.cacheEntry.sizedEntry.Size().Bytes())
	c.evictLocked()

	// Prevent double-releasing.
	r.Value = nil
	r.lruCache = nil
	r.cacheEntry = nil
}

func (c *LRUCache) EntryCount() int {
	return len(c.mapping)
}

func (c *LRUCache) Size() Byte {
	return c.totalSize
}

func (c *LRUCache) EvictableSize() Byte {
	return c.evictableSize
}

func (c *LRUCache) OvercommittedSize() Byte {
	return MaxBytes(
		Byte(0),
		Byte(c.totalSize.Bytes()-c.sizeLimit.Bytes()),
	)
}
