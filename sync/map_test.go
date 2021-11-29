package sync

import (
	"sync"
	"testing"
)

/**
tips: go map 并发并不安全。
所以，如果需要并发安全有两个策略：
	1. 读写锁 + map
	2. sync.Map
sync.Map底层实现了两个map，一个read map，一个dirty map
写操作是直接写dirty，读操作首先读read，如果未命中则读取dirty中。
本质上是一种，用空间换时间的方法。
通过基线测试，如果只是单纯的写操作，并发较少的情况下，读写锁 + map 优于 sync.Map，但是 sync.Map 的空间使用往往是前者的好几倍（大于2）
*/

type lockMap struct {
	sync.RWMutex
	m map[string]int
}

func (lm *lockMap) read(key string) int {
	lm.RLock()
	defer lm.RUnlock()
	return lm.m[key]
}

func (lm *lockMap) do(N int, key string, wg *sync.WaitGroup) {
	for i := 0; i < N; i++ {
		// 加锁
		lm.Lock()
		lm.m[key] = 1
		// 解锁
		lm.Unlock()
	}
	wg.Done()
}

func do(m *sync.Map, N int, key string, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < N; i++ {
		m.Store(key, 1)
		i++
	}
}

func BenchmarkLockMapWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := new(sync.WaitGroup)
		lm := &lockMap{
			m: map[string]int{},
		}
		key := "test"
		// 两个协程争夺锁
		for n := 0; n < 500; n++ {
			wg.Add(1)
			go lm.do(100, key, wg)
		}
		wg.Wait()
	}
}

func BenchmarkSyncMapWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := new(sync.WaitGroup)
		key := "test"
		lm := new(sync.Map)
		for n := 0; n < 500; n++ {
			wg.Add(1)
			go do(lm, 100, key, wg)
		}
		wg.Wait()
	}
}

func BenchmarkLockMapWriteAndRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := new(sync.WaitGroup)
		lm := &lockMap{
			m: map[string]int{},
		}
		key := "test"
		// 两个协程争夺锁
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go lm.do(100, key, wg)
		}
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go func() {
				lm.read(key)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkSyncMapWriteAndRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wg := new(sync.WaitGroup)
		key := "test"
		lm := new(sync.Map)
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go do(lm, 100, key, wg)
		}
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			go func() {
				lm.Load(key)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
