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
type SizedEntryRef[T SizedEntry] struct {
	Value      T
	lruCache   *LRUCache[T]
	cacheEntry *lruCacheEntry[T]
}

// A SizedEntryFactory is a factory that can create a SizedEntry given its key
// name.
type SizedEntryFactory[T SizedEntry] func(key string) (T, error)

// An lruCacheEntry is an entry into an LRUCache.
//
// refCount is zero if and only if listElement is non-nil.
type lruCacheEntry[T SizedEntry] struct {
	refCount    int32
	listElement *list.Element
	sizedEntry  T
	key         string
}

// LRUCache handles a pool of sized resources. It has a fixed maximum size with
// a least-recently used eviction policy.
type LRUCache[T SizedEntry] struct {
	sync.Mutex
	mapping       map[string]*lruCacheEntry[T]
	evictList     *list.List
	totalSize     Byte
	evictableSize Byte
	sizeLimit     Byte
}

// NewLRUCache returns an empty LRUCache with the provided size limit.
func NewLRUCache[T SizedEntry](sizeLimit Byte) *LRUCache[T] {
	return &LRUCache[T]{
		mapping:   make(map[string]*lruCacheEntry[T]),
		evictList: list.New(),
		sizeLimit: sizeLimit,
	}
}

func (c *LRUCache[T]) evictLocked() {
	for c.evictList.Len() > 0 && c.totalSize.Bytes() > c.sizeLimit.Bytes() {
		element := c.evictList.Back()
		cacheEntry := element.Value.(*lruCacheEntry[T])

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

func (c *LRUCache[T]) reserveLocked(size Byte) {
	c.totalSize = Byte(c.totalSize.Bytes() + size.Bytes())
	c.evictLocked()
}

// Get atomically gets a previously-created entry if it was found in the cache,
// or a newly-created one otherwise. It is the caller's responsibility to call
// Put() with the returned SizedEntryRef method once it's no longer needed so
// that the underlying resource can be evicted from the cache, if needed.
func (c *LRUCache[T]) Get(
	key string,
	factory SizedEntryFactory[T],
) (*SizedEntryRef[T], error) {
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
		return &SizedEntryRef[T]{
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
	cacheEntry := &lruCacheEntry[T]{
		refCount:   1,
		sizedEntry: value,
		key:        key,
	}

	c.mapping[key] = cacheEntry

	return &SizedEntryRef[T]{
		Value:      value,
		lruCache:   c,
		cacheEntry: cacheEntry,
	}, nil
}

// Put marks a SizedEntryRef as no longer being referred to, so that it can be
// considered for eviction.
func (c *LRUCache[T]) Put(r *SizedEntryRef[T]) {
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
	var zero T
	r.Value = zero
	r.lruCache = nil
	r.cacheEntry = nil
}

// EntryCount is the number of elements in the LRUCache.
func (c *LRUCache[T]) EntryCount() int {
	return len(c.mapping)
}

// Size is the total size in bytes of all the elements in the LRUCache.
func (c *LRUCache[T]) Size() Byte {
	return c.totalSize
}

// EvictableSize is the size in bytes of all elements that are being considered
// for eviction. This is, not currently being used.
func (c *LRUCache[T]) EvictableSize() Byte {
	return c.evictableSize
}

// OvercommittedSize is the size in bytes that have been allocated above the
// LRUCache's size limit. This number can be non-zero when all the elements in
// the cache are currently being used and cannot yet be evicted.
func (c *LRUCache[T]) OvercommittedSize() Byte {
	return Max(
		Byte(0),
		Byte(c.totalSize.Bytes()-c.sizeLimit.Bytes()),
	)
}
