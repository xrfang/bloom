package bloom

import (
	_ "crypto/sha1"
	"encoding/binary"
	"hash/crc64"
	"math"
)

const BUF_SIZE = 8192

type bitmap [BUF_SIZE]byte

type Filter struct {
	buffers   []bitmap
	count     int
	hashCount int
	tbl       *crc64.Table
}

func NewFilter(capacity int, falsePositive float64) *Filter {
	entropy := math.Log(falsePositive) / math.Log(0.6185)
	size := float64(capacity) * entropy / 8
	hc := int(-math.Log(falsePositive)/math.Log(2) + 0.5)
	return &Filter{
		buffers:   make([]bitmap, int(size/BUF_SIZE+0.5)),
		hashCount: hc,
		tbl:       crc64.MakeTable(crc64.ECMA),
	}
}

func (f *Filter) hash(item []byte, seq int) []uint {
	var hashs []uint
	var i byte
	w := int(math.Log2(BUF_SIZE*8) / 8) //width of buffer in bytes
	chunk := 8 / w
	for {
		i++
		sum := make([]byte, 8)
		//sum := sha1.Sum(append(item, byte(seq), i))
		data := crc64.Checksum(append(item, byte(seq), i), f.tbl)
		binary.LittleEndian.PutUint64(sum, data)
		for j := 0; j < chunk; j++ {
			var idx []byte
			for k := 0; k < w; k++ {
				idx = append(idx, sum[j*2+k])
			}
			idx = append(idx, 0, 0, 0, 0, 0, 0, 0)
			hashs = append(hashs, uint(binary.LittleEndian.Uint64(idx)))
		}
		if len(hashs) >= f.hashCount {
			break
		}
	}
	return hashs[:f.hashCount]
}

func (f *Filter) Add(item []byte) {
	for i := 0; i < len(f.buffers); i++ {
		hashs := f.hash(item, i)
		for _, h := range hashs {
			f.buffers[i][h/8] = f.buffers[i][h/8] | (1 << (h % 8))
		}
	}
	f.count++
}

func (f *Filter) Contains(item []byte) bool {
	for i := 0; i < len(f.buffers); i++ {
		hashs := f.hash(item, i)
		for _, h := range hashs {
			if f.buffers[i][h/8]&(1<<(h%8)) == 0 {
				return false
			}
		}
	}
	return true
}

func (f *Filter) Count() int {
	return f.count
}

func (f *Filter) FalsePositive() float64 {
	return math.Pow(1-math.Exp(float64(-f.hashCount*f.count)/BUF_SIZE/8),
		float64(len(f.buffers)*f.hashCount))
}

func (f *Filter) Size() int {
	return BUF_SIZE * len(f.buffers)
}
