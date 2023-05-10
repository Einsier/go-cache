package lru

import (
	"reflect"
	"testing"
)

type String string

// implement Value interface
func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	c := New(int64(0), nil) // zero maxBytes means no limit
	c.Add("key1", String("1234"))
	if v, ok := c.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := c.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "123", "458", "789"
	c := New(int64(len(k1+v1+k2+v2)), nil)
	c.Add(k1, String(v1))
	c.Add(k2, String(v2))
	c.Add(k3, String(v3))
	if _, ok := c.Get("key1"); ok || c.Len() != 2 {
		t.Fatalf("RemoveOldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	c := New(int64(10), callback)
	c.Add("key1", String("1234"))
	c.Add("k2", String("v2"))
	c.Add("k3", String("v3"))
	c.Add("k4", String("v4"))
	if _, ok := c.Get("key1"); ok || c.Len() != 2 {
		t.Fatalf("RemoveOldest key1 and k2 failed")
	}
	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
