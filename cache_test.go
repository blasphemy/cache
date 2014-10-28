package cache

import "testing"
import "time"

var (
	ok *Cache
)

func TestTest(t *testing.T) {
	toplel := CacheOptions{}
	toplel.CacheStrategy = CacheStrategyOldest
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
	op.CacheStrategy = CacheStrategyRandom
	op.MaxEntries = 0
	op.Upper = 0
	op.ExpirationTime = time.Second * 2
	oj := NewCache(op)
	oj.Set("Tests", "FOO BAR")
	if oj.Get("Tests") != "FOO BAR" {
		t.Error("Setup for TestExpireTime failed")
		return
	}
	oj.Start()
	time.Sleep(time.Second * 3)
	if oj.Get("Tests") != nil {
		t.Error("TestExpireTime failed")
	}
	oj.Stop()
}

func TestExpireDead(t *testing.T) {
	op := CacheOptions{}
	op.CacheStrategy = CacheStrategyRandom
	op.MaxEntries = 0
	op.Upper = 0
	op.ExpirationTime = time.Second * 2
	oj := NewCache(op)
	oj.Set("Tests", "FOO BAR")
	if oj.Get("Tests") != "FOO BAR" {
		t.Error("Setup for TestExpireTime failed")
		return
	}
	oj.Start()
	oj.Stop()
	time.Sleep(time.Second * 3)
	if oj.Get("Tests") != "FOO BAR" {
		t.Fail()
	}
	oj.Stop()
}

func TestExpireNoDuration(t *testing.T) {
	op := CacheOptions{}
	op.CacheStrategy = CacheStrategyRandom
	op.MaxEntries = 0
	op.Upper = 0
	oj := NewCache(op)
	oj.Set("Tests", "FOO BAR")
	if oj.Get("Tests") != "FOO BAR" {
		t.Error("Setup for TestExpireTime failed")
		return
	}
	oj.Start()
	time.Sleep(time.Second * 3)
	if oj.Get("Tests") != "FOO BAR" {
		t.Fail()
	}
	oj.Stop()
}

func TestHits(t *testing.T) {
	toplel := CacheOptions{}
	toplel.CacheStrategy = CacheStrategyOldest
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
	toplel.CacheStrategy = CacheStrategyOldest
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
	toplel.CacheStrategy = CacheStrategyOldest
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

func TestBurnEntryByRandom(t *testing.T) {
	toplel := CacheOptions{}
	toplel.CacheStrategy = CacheStrategyRandom
	toplel.Upper = 500
	ok = NewCache(toplel)
	ok.Set("first", 1)
	ok.Set("Second", 2)
	if ok.Len() != 2 {
		t.Fail()
		return
	}
	ok.Trim(1)
	if ok.Len() != 1 {
		t.Error("Len ", ok.Len())
	}
}

func TestDoubleSetValue(t *testing.T) {
	toplel := CacheOptions{}
	toplel.Upper = 500
	ok = NewCache(toplel)
	ok.Set("first", 1)
	ok.Set("Second", 2)
	ok.Set("first", 3)
	if ok.l.Len() != 2 || len(ok.contents) != 2 {
		t.Error("MAP LEN ", len(ok.contents))
		t.Error("LIST LEN", ok.Len())
		return
	}
}
