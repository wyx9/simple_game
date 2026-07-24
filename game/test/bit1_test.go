package test_test

import (
	"fmt"
	"testing"
)

func Test_bit1(t *testing.T) {
	fmt.Printf("%08b\n", 0)
	fmt.Printf("%08b\n", 1)

	// &  |
	fmt.Printf("%08b\n", 5)
	fmt.Printf("%08b\n", 3)
	fmt.Println(5 | 3)
	fmt.Println("-----------------")
	fmt.Printf("%08b\n", 7)

	// 位运算
	// 左移
	fmt.Println("-----------------")
	fmt.Printf("%08b\n", 5)
	// 相当于*2
	fmt.Printf("%08b\n", 5<<1)
	fmt.Printf("%08b\n", 5<<4)

	// 右移
	fmt.Println("-----------------")
	fmt.Printf("%08b\n", 5)
	// 相当于 /2
	fmt.Printf("%08b\n", 5>>2) //1

	// 取二进制个数的最末尾
	fmt.Println("-----------------")
	fmt.Printf("%08b\n", 5)
	fmt.Println(5 & 1)

}

func Test_bit2(t *testing.T) {
	fmt.Printf("%08b\n", 1<<1)
	fmt.Printf("%08b\n", 0|1<<1)

	fmt.Printf("%08b\n", 1<<3)
	fmt.Printf("%08b\n", 0|1<<3)

	// 两个32位合并
	// int64(source)<<32 | int64(targetID)
	// 64位解析 两个32
	temp := []int32{312312, 12312444, 3}
	a := int64(temp[0])<<32 | int64(temp[1])
	// 取高32
	ena := a >> 32
	// 取低32
	enb := a & 0xFFFFFFFF
	fmt.Printf("%08b\n", a)
	fmt.Printf("%08b\n", 0xFFFFFFFF)
	fmt.Printf("%08b\n", a&0xFFFFFFFF)

	fmt.Println(a)
	fmt.Println(ena)
	fmt.Println(enb)

	fmt.Printf("%08b\n", 150001231200)
	param1 := 123 << 32
	fmt.Printf("%08b\n", param1)

	param2 := (150001231200 << 32) >> 32
	fmt.Printf("%08b\n", param2)
	fmt.Println(param2)

	fmt.Println("--------------------------------------------------------")
	fmt.Printf("%08b\n", int32(1999))
	fmt.Printf("%08b\n", int64(1999))
}
