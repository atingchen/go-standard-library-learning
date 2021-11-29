package sync

import (
	"bytes"
	"sync"
	"testing"
)

/**
	tips: sync.Pool 可以复用已有对象，同时支持可伸缩，减少GC压力。
	tps-telemetry 指标聚合Processor，需要根据resource等结构体的信息，生成一个string类型的key，用到了byte.Buffer。
	但是如果每次生成key，都使用一个新的byte.Buffer，则会在堆上创建大量的`临时变量`，引起较多的GC

     go test -bench . -benchmem

							基准测试的迭代总次数	平均每次迭代所消耗的纳秒数  平均每次迭代内存所分配的字节数 平均每次迭代的内存分配次数
BenchmarkWrite-12               18170775                58.33 ns/op              224 B/op          			1 allocs/op
BenchmarkWriteWithPool-12       65204690                16.80 ns/op                0 B/op          			0 allocs/op

*/

var data = []byte("hello world!")

func BenchmarkWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(make([]byte, 100))
		buf.Write(data)
	}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 100))
	},
}

func BenchmarkWriteWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		buf.Write(data)
		bufferPool.Put(buf)
	}
}
