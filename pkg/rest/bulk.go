package rest

import (
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type fakeDoer struct{}

func (f fakeDoer) Do(a []any) error {
	return nil
}

type ServiceLayer interface {
	Do([]any) error
}

type BulkProcessor struct {
	mu sync.RWMutex
	mm map[uint64]Shard

	svc ServiceLayer
}

func NewBulkProcessor() *BulkProcessor {
	return &BulkProcessor{
		mm:  make(map[uint64]Shard),
		svc: &fakeDoer{},
	}
}

func (bp *BulkProcessor) DoStuff(c echo.Context) error {

	shard, running, clean := bp.FindShard(50)

	if !running {
		go shard.BulkRequests(10*time.Millisecond, 150, clean)
	}

	shard.ch <- 1

	return nil

}

type Shard struct {
	ch  chan int
	svc ServiceLayer
}

func (s Shard) BulkRequests(timeout time.Duration, maxSize int, fnClose func()) {
	var reqs []any

	tt := time.NewTimer(timeout)
	defer func() {
		if !tt.Stop() {
			<-tt.C
		}
	}()

loop:
	for {
		select {
		case req, ok := <-s.ch:
			if !ok {
				if len(reqs) > 0 {
					s.svc.Do(reqs)
				}
				break loop
			}
			reqs = append(reqs, req)
			if len(reqs) >= maxSize {
				fnClose()
				s.svc.Do(reqs)
				break loop
			}
		case <-time.After(timeout):
			fnClose()
			s.svc.Do(reqs)
			break loop
		}
	}
}

func (bp *BulkProcessor) FindShard(id uint64) (ch Shard, running bool, clean func()) {
	bp.mu.RLock()
	if ch, ok := bp.mm[id]; ok {
		bp.mu.RUnlock()
		return ch, true, func() {
			bp.mu.Lock()
			defer bp.mu.Unlock()
			close(bp.mm[id].ch)
			delete(bp.mm, id)
		}
	}
	bp.mu.RUnlock()
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.mm[id] = Shard{
		ch:  make(chan int, 1),
		svc: bp.svc,
	}
	return ch, false, func() {
		bp.mu.Lock()
		defer bp.mu.Unlock()
		close(bp.mm[id].ch)
		delete(bp.mm, id)
	}
}

type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewSafeMap[K comparable, V any](size int) *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V, size),
	}
}

func (sm *SafeMap[K, V]) Get(k K) (v V, ok bool) {
	sm.mu.RLock()
	v, ok = sm.m[k]
	sm.mu.RUnlock()
	return
}

func (sm *SafeMap[K, V]) Set(k K, v V) {
	sm.mu.Lock()
	sm.m[k] = v
	sm.mu.Unlock()
}

func (sm *SafeMap[K, V]) Delete(k K) {
	sm.mu.Lock()
	delete(sm.m, k)
	sm.mu.Unlock()
}

func (sm *SafeMap[K, V]) GetOrSet1(k K, v V) (V, bool) {
	sm.mu.RLock()
	val, ok := sm.m[k]
	if ok {
		sm.mu.RUnlock()
		return val, ok
	}
	sm.mu.RUnlock()
	sm.mu.Lock()
	sm.m[k] = v
	sm.mu.Unlock()
	return v, true
}

func (sm *SafeMap[K, V]) GetOrSet2(k K, v V) (V, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	val, ok := sm.m[k]
	if !ok {
		sm.m[k] = v
		return val, ok
	}
	return val, ok
}
