package test_test

import (
	"fmt"
	"math/rand"
	"testing"
)

func Test_Byte1(t *testing.T) {
	//var b1 byte = 255
	//
	//a := 0 + float32(1+1.3)
	//fmt.Print(int32(a))
	//a := false
	//
	//go func() {
	//	time.Sleep(time.Second * 3)
	//	a = true
	//}()
	//
	//for {
	//	if a == true {
	//		continue
	//	}
	//}

	for i := 0; i < 100; i++ {
		fmt.Println(rand.Int31n(2))
	}
}
