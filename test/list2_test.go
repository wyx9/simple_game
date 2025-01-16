package test_test

import (
	"fmt"
	"testing"
	"time"
)

func Test_2(t *testing.T) {
	list := make([]int, 0)
	for i := 0; i < 200000000; i++ {
		list = Add(list, i)
	}
	fmt.Print(list)

}
func Test_3(t *testing.T) {
	list := make([]int, 0)
	for i := 0; i < 200000000; i++ {
		Add3(&list, i)
	}
	fmt.Print(list)
}

func Add(list []int, a int) []int {
	temp := make([]int, 0)
	if len(list) >= 50 {
		temp = list[1:]
		temp = append(temp, a)
	} else {
		temp = append(list, a)
	}
	return temp
}

func Add2(list []int) {
	list[0] = 12312312
}

// 队列？？
func Add3(list *[]int, a int) {
	if len(*list) >= 50 {
		temp := make([]int, 0)
		ints := *list
		temp = ints[1:]
		*list = temp
	}
	*list = append(*list, a)

}

func Test_4(t *testing.T) {
	list := make([]int, 0)
	for i := 0; i < 48; i++ {
		list = Add(list, i)
	}
	fmt.Println(list)

	//offset := 36
	//limit := 12
	//ints := list[offset : limit+offset]
	fmt.Println(len(list))
	ints := list[48:]
	fmt.Println(ints)

	// make(数组，长度，容量) 一般来说，容量初始设定后对于以后append 扩容效率有所影响
	//list2 := make([]int32, 10, 11)
	//fmt.Println(list2)
	//list2[0] = 1
	weekday := time.Now().Unix()
	for i := 0; i < 7; i++ {
		weekday += 86400
		w := time.Unix(weekday, 0).Weekday()
		fmt.Println("----------", int(w))
	}

}
