package gkv

import (
	"context"
	"sync"
	"time"
)

type Gkv struct {
	containerMu sync.RWMutex
	container   map[string]*Value
	ctx         context.Context
	cancel      context.CancelFunc
}

func New() *Gkv {
	g := &Gkv{}
	g.container = map[string]*Value{}
	g.ctx, g.cancel = context.WithCancel(context.Background())
	go g.cleanTask()
	return g
}
func (g *Gkv) Close() error {
	g.cancel()
	return nil
}
func (g *Gkv) Len() int {
	g.containerMu.RLock()
	defer g.containerMu.RUnlock()
	return len(g.container)
}

func (g *Gkv) Set(key string, value interface{}, expireAt *time.Time) {
	g.containerMu.Lock()
	defer g.containerMu.Unlock()
	v := &Value{}
	v.expireAt = expireAt
	v.Val = value
	g.container[key] = v
}
func (g *Gkv) Get(key string) *Value {
	g.containerMu.RLock()
	defer g.containerMu.RUnlock()
	if v := g.getWithoutLock(key); v != nil {
		return v
	}
	return nil
}

func (g *Gkv) getWithoutLock(key string) *Value {
	if v, ok := g.container[key]; ok {
		if v.expireAt != nil {
			if v.expireAt.After(time.Now()) {
				return v
			}
		} else {
			return v
		}
	}
	return nil
}

func (g *Gkv) GetOrCreate(key string, val interface{}, expireAt *time.Time) (v *Value) {
	g.containerMu.Lock()
	defer g.containerMu.Unlock()

	if v = g.getWithoutLock(key); v == nil {
		v = &Value{
			Val: val,
		}
		v.expireAt = expireAt
		g.container[key] = v
	}
	return
}
func (g *Gkv) GetCopy() (v map[string]interface{}) {
	g.containerMu.RLock()
	defer g.containerMu.RUnlock()
	v = map[string]interface{}{}
	now := time.Now()

	for k, val := range g.container {
		if !val.isExpired(&now) {
			v[k] = val.Val
		}
	}
	return
}

func (g *Gkv) Del(key string) {
	g.containerMu.Lock()
	defer g.containerMu.Unlock()
	delete(g.container, key)
}

func (g *Gkv) cleanExpire() {
	var keys []string
	var now = time.Now()
	g.containerMu.RLock()
	for k, v := range g.container {
		if v.expireAt != nil {
			if now.After(*v.expireAt) {
				keys = append(keys, k)
			}
		}
	}
	g.containerMu.RUnlock()

	g.containerMu.Lock()
	defer g.containerMu.Unlock()

	for _, k := range keys {
		delete(g.container, k)
	}
}

func (g *Gkv) cleanTask() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-time.After(time.Minute):
			g.cleanExpire()
		}
	}
}
