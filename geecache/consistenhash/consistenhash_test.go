package consistenhash

import (
	"fmt"
	"testing"
)

func TestRandString(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(randString(32))
	}
}

func TestAll(t *testing.T) {
	m := NewMap(5, nil)
	m.Add("1", "2", "3", "4")
	fmt.Println(m.Get("123"))
	fmt.Println(m.Get("asdfa"))
	fmt.Println(m.Get("123"))
	fmt.Println(m.Get("asdfa"))

}
