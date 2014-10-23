package cache

import "testing"

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
