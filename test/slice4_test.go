package test_test

import (
	"fmt"
	"testing"
	"time"
)

type Person struct {
	Name  string
	Score int
}

func testsad() {
	//personList := []Person{
	//	{Name: "zs", Score: 44},
	//	{Name: "zs2", Score: 123},
	//	{Name: "z3", Score: 124},
	//	{Name: "zs4", Score: 2},
	//	{Name: "z5", Score: 143},
	//	{Name: "zs6", Score: 213},
	//	{Name: "zs7", Score: 3},
	//}
	//
	//sort.Slice(personList, func(i, j int) bool {
	//	return personList[i].Score < personList[j].Score
	//})
	//
	//fmt.Println(personList[1:])
	//f := float32(100.0) * float32(1.05)
	//fmt.Println(int32(f))
	//fmt.Println(int32(f))

	//	list := make([]int, 100)
	//	for i := 0; i < 100; i++ {
	//		list[i] = i
	//	}
	//
	//	fmt.Println(list[:12])
	//	fmt.Println(list[12:24]) //   [ )
	//
	//	s := `
	//asd
	//adsasd
	//dasda
	//dasdas
	//asdas
	//asdad`
	//	fmt.Println(s)

	var c = make(chan int32)
	go func() {
		for {
			select {
			case v := <-c: // 检测有没有数据可读
				// 一旦成功读取到数据，则进行该case处理语句
				fmt.Printf("写入 %d", v)
			case c <- 100: // 检测有没有数据可写
				// 一旦成功向c写入数据，则进行该case处理语句
				fmt.Printf("写出")
			default:
				// 如果以上都没有符合条件，那么进入default处理流程
			}
		}
	}()
	time.Sleep(1000)
	c <- 100
}

func Test_222(t *testing.T) {
	//testing.AllocsPerRun(1, testsad)

	//for i := 4; i <= 5; i++ {
	//	fmt.Println("asdasdsa")
	//}

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

	fmt.Printf("%08b\n", 2)

}
