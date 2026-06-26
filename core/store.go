package core

import (
	"time"
)

var store = make(map[string]*Obj)

type Obj struct {
	Value     interface{}
	ExpiresAt int64 // Unix timestamp in seconds, -1 if it doesn't expire
}

func NewObj(value interface{}, durationMs int64) *Obj {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}

	return &Obj{
		Value:     value,
		ExpiresAt: expiresAt,
	}

}

func Set(key string, value *Obj) {
	store[key] = value
}

func Get(key string) (*Obj, bool) {
	obj, exists := store[key]
	if !exists {
		return nil, false
	}

	if obj.ExpiresAt != -1 && time.Now().UnixMilli() > obj.ExpiresAt {
		// Todo: Remove the key from the store if it has expired
		return nil, false
	}

	return obj, true
}

func Delete(key string) {
	delete(store, key)
}
