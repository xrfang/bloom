package bloom

import (
	"fmt"
	"testing"
)

func TestBloom(t *testing.T) {
	f := NewFilter(0.005)
	fmt.Printf("%v\n", f.Add([]byte("hello")))
	fmt.Printf("%v\n", f.Add([]byte("hello")))
	fmt.Printf("%v\n", f.Add([]byte("world")))
	fmt.Printf("%v\n", f.Add([]byte("world")))
}
