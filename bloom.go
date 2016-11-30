package bloom

import (
	"hash"
	"hash/crc64"
	. "math"
)

type Filter struct {
	blocks []*bitmap
	blkcap int
	blkidx int
}

func NewFilter(falsePositive float64) *Filter {
	f := Filter{}
	f.blocks = append(f.blocks, newBitmap())
	f.blkcap = int(256 * Log(0.6185) / Log(falsePositive))
	return &f
}

func (f *Filter) Blocks() int {
	return len(f.blocks)
}

func (f *Filter) Clear(cnt int) {
	if cnt > 0 && cnt < len(f.blocks) {
		f.blocks = f.blocks[cnt:]
		return
	}
	f.blocks = []*bitmap{f.blocks[len(f.blocks)-1]}
}

func (f *Filter) Items() int {
	idx := len(f.blocks) - 1
	return idx*f.blkcap + f.blocks[idx].cnt
}

func (f *Filter) FalsePositive() float64 {
	f0 := f.blocks[0].falsePositive()
	n := float64(len(f.blocks))
	if n > 1 {
		return ((n-1)*f0 + f.blocks[len(f.blocks)-1].falsePositive()) / n
	}
	return f0
}

func (f *Filter) Contains(item []byte) bool {
	sum := f.blocks[0].checksum(item)
	for _, b := range f.blocks {
		if b.contains(sum) {
			return true
		}
	}
	return false
}

func (f *Filter) Add(item []byte) bool {
	sum := f.blocks[0].checksum(item)
	for _, b := range f.blocks {
		if b.contains(sum) {
			return true
		}
	}
	if f.blocks[f.blkidx].cnt >= f.blkcap {
		f.blkidx++
		if f.blkidx >= len(f.blocks) {
			f.blocks = append(f.blocks, newBitmap())
		}
	}
	f.blocks[f.blkidx].add(sum)
	f.blocks[f.blkidx].cnt++
	return false
}

type bitmap struct {
	buf []byte
	h   hash.Hash
	cnt int
}

func newBitmap() *bitmap {
	return &bitmap{
		buf: make([]byte, 32),
		h:   crc64.New(crc64.MakeTable(crc64.ECMA)),
	}
}

func (b *bitmap) falsePositive() float64 {
	return Pow(1-Exp(-8*float64(b.cnt)/256), 8)
}

func (b *bitmap) checksum(item []byte) []byte {
	b.h.Reset()
	return b.h.Sum(item)
}

func (b *bitmap) contains(checksum []byte) bool {
	for _, s := range checksum {
		if b.buf[s/8]&(1<<(s%8)) == 0 {
			return false
		}
	}
	return true
}

func (b *bitmap) add(checksum []byte) int {
	for _, s := range checksum {
		b.buf[s/8] = b.buf[s/8] | (1 << (s % 8))
	}
	b.cnt++
	return b.cnt
}
