package bloom

import (
	"crypto/rand"
	"testing"
)

func TestRandomAdd(t *testing.T) {
	const (
		CAP    = 6000
		LOAD   = 10000
		EXP_FP = 0.0001
	)
	f := NewFilter(CAP, EXP_FP)
	t.Logf("capacity=%d; load=%d; expected fp=%0.4f", CAP, LOAD, EXP_FP)
	fp := 0
	for i := 0; i < LOAD; i++ {
		data := make([]byte, 8)
		_, err := rand.Read(data)
		if err != nil {
			panic(err)
		}
		if f.Contains(data) {
			fp++
		}
		f.Add(data)
		if !f.Contains(data) {
			t.Fatalf("broken bloom filter")
		}
	}
	t.Logf("items=%d; fp(real)=%0.4f; fp(theory)=%0.4f\n", f.Count(),
		float64(fp)/10000, f.FalsePositive())
}
