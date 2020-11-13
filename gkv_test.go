package gkv

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGkv_Set(t *testing.T) {
	c := New()
	var iter uint32
	wait := sync.WaitGroup{}
	const gr = 1000
	wait.Add(gr)

	for i := 0; i < gr; i++ {
		go func() {
			defer wait.Done()
			for j := 0; j < 1000; j++ {
				newI := atomic.AddUint32(&iter, 1)
				c.Set(fmt.Sprintf("%d", newI), newI, nil)
			}
		}()
	}
	wait.Wait()
	wantLen := gr * 1000
	if c.Len() != 1000*1000 {
		t.Errorf("len (%d)not (%d)", c.Len(), wantLen)
	}
}
func TestGkv_Get(t *testing.T) {
	c := New()
	var iter uint32
	const (
		gr       = 1000
		keyCount = 4
	)
	for i := 0; i < keyCount; i++ {
		newI := atomic.AddUint32(&iter, 1)
		c.Set(fmt.Sprintf("%d", newI), newI, nil)
	}

	wait := sync.WaitGroup{}
	wait.Add(gr)

	for i := 0; i < gr; i++ {
		go func() {
			defer wait.Done()
			k := 1
			for j := 0; j < 1000; j++ {
				if k > keyCount {
					k = 1
				}
				v := c.Get(fmt.Sprintf("%d", k))
				v.Lock()
				v.val = time.Now()
				v.val = k + 1
				v.Unlock()

				newI := atomic.AddUint32(&iter, 1)
				c.Set("6", newI, nil)

				k++
			}
			c.Set("6", 3333, nil)
		}()
	}
	wait.Wait()
	wantLen := 5
	if c.Len() != wantLen {
		t.Errorf("len (%d)not (%d)", c.Len(), wantLen)
	}
	for i := 1; i < 5; i++ {
		want := i + 1
		if v := c.Get(fmt.Sprintf("%d", i)).val; v != want {
			t.Errorf("val(%d) not %d", v, want)
		}
	}

	if v := c.Get("6").val; v != 3333 {
		t.Errorf("6 not 3333 :%d", v)
	}
}

func TestGkv_GetOrCreate(t *testing.T) {
	c := New()
	const (
		gr  = 100000
		key = "1"
	)

	for i := 0; i < gr; i++ {
		go func(i int) {
			f := func() {
				v := c.GetOrCreate(key)
				v.Lock()
				defer v.Unlock()
				v.val = i
				v.val = 1
			}
			for {
				f()
				time.Sleep(time.Nanosecond * 10)
			}
		}(i)
	}
	<-time.After(time.Millisecond * 500)

	wantLen := 1
	if c.Len() != wantLen {
		t.Errorf("len (%d)not (%d)", c.Len(), wantLen)
	}

	if v := c.Get("1").val; v != 1 {
		t.Errorf("1 not 1 :%d", v)
	}
}

func TestGkv_cleanExpire(t *testing.T) {
	c := New()
	ex1 := time.Now().Add(-time.Millisecond)
	ex2 := time.Now().Add(time.Millisecond * 50)
	c.Set("1", 1, &ex1)
	c.Set("2", 2, &ex2)
	c.cleanExpire()
	if c.Get("1") != nil {
		t.Errorf("1 exist")
	}
	if c.Get("2") == nil {
		t.Errorf("2 not exist")
	}
}
