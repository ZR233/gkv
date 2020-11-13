package gkv

import (
	"sync"
	"time"
)

type Value struct {
	val      interface{}
	expireAt *time.Time
	sync.RWMutex
}

func (v *Value) Val() interface{} {
	return v.val
}
func (v *Value) ExpireAt(at time.Time) {
	v.expireAt = &at
}
