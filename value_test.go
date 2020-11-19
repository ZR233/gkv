package gkv

import (
	"sync"
	"testing"
	"time"
)

func TestValue_isExpired(t *testing.T) {
	type fields struct {
		val      interface{}
		expireAt *time.Time
		RWMutex  sync.RWMutex
	}
	type args struct {
		now *time.Time
	}

	t1 := time.Now()
	t2 := t1.Add(-time.Second)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"1", fields{
			val:      nil,
			expireAt: &t2,
		}, args{&t1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{
				val:      tt.fields.val,
				expireAt: tt.fields.expireAt,
				RWMutex:  tt.fields.RWMutex,
			}
			if got := v.isExpired(tt.args.now); got != tt.want {
				t.Errorf("isExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
