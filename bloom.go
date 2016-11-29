package bloom

import (
	"crypto/sha1"
	. "math"
)

type Filter struct {
	blocks []*bitmap
	fpRate float64
}

func NewFilter(falsePositive float64) *Filter {
	f := Filter{fpRate: falsePositive}
	f.blocks = append(f.blocks, newBitmap())
	return &f
}

func (f *Filter) Contains(item []byte) bool {
	for _, b := range f.blocks {
		if b.contains(item) {
			return true
		}
	}
	return false
}

func (f *Filter) Add(item []byte) bool {
	exists := f.Contains(item)
	if exists {
		return true
	}
	added := false
	for _, b := range f.blocks {
		if b.falsePositive() >= f.fpRate {
			continue
		}
		b.add(item)
		added = true
	}
	if !added {
		newb := newBitmap()
		newb.add(item)
		f.blocks = append(f.blocks, newb)
	}
	return false
}

func (f *Filter) Blocks() int {
	return len(f.blocks)
}

type bitmap struct {
	buf []byte
	cnt int
}

func newBitmap() *bitmap {
	return &bitmap{buf: make([]byte, 8192)}
}

func (b *bitmap) capacity(falsePositive float64) int {
	return int(65536 * Log(0.6185) / Log(falsePositive))
}

func (b *bitmap) falsePositive() float64 {
	return Pow(1-Exp(-10*float64(b.cnt)/65536), 10)
}

func (b *bitmap) clear() {
	b.buf = make([]byte, 8192)
	b.cnt = 0
}

func (b *bitmap) hash(item []byte) (idx [10]uint16) {
	sum := sha1.Sum(item)
	for i := 0; i < 10; i++ {
		idx[i] = uint16(sum[i*2])<<8 + uint16(sum[i*2+1])
	}
	return
}

func (b *bitmap) contains(item []byte) bool {
	for _, h := range b.hash(item) {
		if b.buf[h/8]&(1<<(h%8)) == 0 {
			return false
		}
	}
	return true
}

func (b *bitmap) add(item []byte) bool {
	exists := b.contains(item)
	if exists {
		return true
	}
	for _, h := range b.hash(item) {
		b.buf[h/8] = b.buf[h/8] | (1 << (h % 8))
	}
	b.cnt++
	return false
}
