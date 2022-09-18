package base

import (
	"fmt"
	"testing"
)

type releasable struct {
	released bool
	size     Byte
	t        *testing.T
}

func (r *releasable) Release() {
	fmt.Printf("Releasing %p\n", r)
	if r.released {
		r.t.Fatalf("releasable object released more than once")
	}
	r.released = true
}

func (r *releasable) Size() Byte {
	return r.size
}

func (r *releasable) Value() *releasable {
	return r
}

func TestLRUCache(t *testing.T) {
	c := NewLRUCache[*releasable](Kibibyte)

	if c.Size().Bytes() != 0 {
		t.Fatalf("c.Size() = %d; want 0", c.Size().Bytes())
	}

	r := releasable{
		size: Byte(1),
	}
	r1k := releasable{
		size: Kibibyte,
	}
	r1M := releasable{
		size: Mebibyte,
	}

	// New object is created.
	{
		created := false
		ref, err := c.Get("r", func(key string) (SizedEntry[*releasable], error) {
			created = true
			return &r, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !created {
			t.Errorf("expected an object to be created, and wasn't")
		}
		if ref.Value != &r {
			t.Errorf("ref.Value = %p; want %p", ref.Value, &r)
		}

		if c.Size().Bytes() != 1 {
			t.Fatalf("c.Size() = %d; want 1", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 0 {
			t.Fatalf("c.EvictableSize() = %d; want 0", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r.released {
			t.Fatalf("releasable object unexpectedly released")
		}

		c.Put(ref)

		if c.Size().Bytes() != 1 {
			t.Fatalf("c.Size() = %d; want 1", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 1 {
			t.Fatalf("c.EvictableSize() = %d; want 1", c.EvictableSize().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r.released {
			t.Fatalf("releasable object unexpectedly released")
		}
	}

	// Object is reused.
	{
		created := false
		ref, err := c.Get("r", func(key string) (SizedEntry[*releasable], error) {
			created = true
			return &r, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if created {
			t.Errorf("expected an object to be reused, and wasn't")
		}
		if ref.Value != &r {
			t.Errorf("ref.Value = %p; want %p", ref.Value, &r)
		}

		if c.Size().Bytes() != 1 {
			t.Fatalf("c.Size() = %d; want 1", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 0 {
			t.Fatalf("c.EvictableSize() = %d; want 0", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r.released {
			t.Fatalf("releasable object unexpectedly released")
		}

		c.Put(ref)

		if c.Size().Bytes() != 1 {
			t.Fatalf("c.Size() = %d; want 1", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 1 {
			t.Fatalf("c.EvictableSize() = %d; want 1", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r.released {
			t.Fatalf("releasable object unexpectedly released")
		}
	}

	// New object is created, old one is evicted.
	{
		created := false
		ref, err := c.Get("r1k", func(key string) (SizedEntry[*releasable], error) {
			created = true
			return &r1k, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !created {
			t.Errorf("expected an object to be created, and wasn't")
		}
		if !r.released {
			t.Fatalf("releasable object %p was not released", &r)
		}
		if ref.Value != &r1k {
			t.Errorf("ref.Value = %p; want %p", ref.Value, &r1k)
		}

		if c.Size().Bytes() != 1024 {
			t.Fatalf("c.Size() = %d; want 1024", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 0 {
			t.Fatalf("c.EvictableSize() = %d; want 0", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r1k.released {
			t.Fatalf("releasable object unexpectedly released")
		}

		c.Put(ref)

		if c.Size().Bytes() != 1024 {
			t.Fatalf("c.Size() = %d; want 1024", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 1024 {
			t.Fatalf("c.EvictableSize() = %d; want 1024", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if r1k.released {
			t.Fatalf("releasable object unexpectedly released")
		}
	}

	// New object is created, and evicted immediately upon release.
	{
		created := false
		ref, err := c.Get("r1M", func(key string) (SizedEntry[*releasable], error) {
			created = true
			return &r1M, nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !created {
			t.Errorf("expected an object to be created, and wasn't")
		}
		if !r1k.released {
			t.Fatalf("releasable object was not released")
		}
		if ref.Value != &r1M {
			t.Errorf("ref.Value = %p; want %p", ref.Value, &r1M)
		}

		if c.Size().Bytes() != 1048576 {
			t.Fatalf("c.Size() = %d; want 1048576", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 0 {
			t.Fatalf("c.EvictableSize() = %d; want 0", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 1047552 {
			t.Fatalf("c.OvercommittedSize() = %d; want 1047552", c.Size().Bytes())
		}
		if r1M.released {
			t.Fatalf("releasable object unexpectedly released")
		}

		c.Put(ref)

		if c.Size().Bytes() != 0 {
			t.Fatalf("c.Size() = %d; want 0", c.Size().Bytes())
		}
		if c.EvictableSize().Bytes() != 0 {
			t.Fatalf("c.EvictableSize() = %d; want 0", c.Size().Bytes())
		}
		if c.OvercommittedSize().Bytes() != 0 {
			t.Fatalf("c.OvercommittedSize() = %d; want 0", c.Size().Bytes())
		}
		if !r1M.released {
			t.Fatalf("releasable object was not released")
		}
	}
}
