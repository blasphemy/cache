package cache

import "testing"
import "time"

var (
	ok *Cache
)

func TestTest(t *testing.T) {
	toplel := CacheOptions{}
	toplel.BurnStrategy = BurnStrategyOldest
	toplel.MaxEntries = 2
	toplel.Upper = 2
	ok = NewCache(toplel)
	ok.Set("Test", "FOO BAR")
	if ok.Get("Test") != "FOO BAR" {
		t.Error("TestTest Failed")
	}
}

func TestRemove(t *testing.T) {
	ok.RemoveItem("Test")
	if ok.Get("Test") != nil || ok.Len() != 0 {
		t.Error("TestRemove Failed")
	}
}

func TestSetNilRemove(t *testing.T) {
	ok.Set("Test", "FOO BAR")
	if ok.Get("Test") != "FOO BAR" {
		t.Error("Setup for TestSetNilRemove Failed")
	}
	ok.Set("Test", nil)
	if ok.Len() != 0 {
		t.Error("TestSetNilRemove failed: did not remove item")
	}
}

func TestExpireTime(t *testing.T) {
	op := CacheOptions{}
	op.BurnStrategy = BurnStrategyRandom
	op.MaxEntries = 0
	op.Upper = 0
	op.ExpirationTime = time.Second * 2
	oj := NewCache(op)
	oj.Set("Tests", "FOO BAR")
	if oj.Get("Tests") != "FOO BAR" {
		t.Error("Setup for TestExpireTime failed")
		return
	}
	time.Sleep(time.Second * 3)
	if oj.Get("Tests") != nil {
		t.Error("TestExpireTime failed")
	}
}

func TestHits(t *testing.T) {
	toplel := CacheOptions{}
	toplel.BurnStrategy = BurnStrategyOldest
	toplel.MaxEntries = 2
	toplel.Upper = 2
	ok = NewCache(toplel)
	ok.Set("test", "lol")
	if ok.Get("test") != "lol" {
		t.Error("Setup")
		return
	}
	if ok.Hits() != 1 {
		t.Error("Fail")
	}
}

func TestMisses(t *testing.T) {
	if ok.Get("non") != nil {
		t.Error("Setup")
		return
	}
	if ok.Misses() != 1 {
		t.Error("Fail")
	}
}

func TestAutoBurnOnUpper(t *testing.T) {
	toplel := CacheOptions{}
	toplel.BurnStrategy = BurnStrategyOldest
	toplel.MaxEntries = 2
	toplel.Upper = 1
	ok = NewCache(toplel)
	ok.Set("first", 1)
	if ok.Get("first") != 1 {
		t.Error("setup")
	}
	ok.Set("second", 2)
	if ok.Get("first") != nil {
		t.Error("first not nil")
		return
	}
	if ok.Get("second") != 2 {
		t.Error("second not 2")
	}
}

func TestBump(t *testing.T) {
	toplel := CacheOptions{}
	toplel.BurnStrategy = BurnStrategyOldest
	toplel.Upper = 500
	ok = NewCache(toplel)
	ok.Set("first", 1)
	ok.Set("Second", 2)
	ok.Bump("first")
	if ok.l.Front().Value.(*CachedItem).key != "first" {
		t.Error("Expected first got", ok.l.Front().Value.(*CachedItem).key)
	}
	if ok.l.Back().Value.(*CachedItem).key != "Second" {
		t.Error("Expected first got", ok.l.Back().Value.(*CachedItem).key)
	}
}
