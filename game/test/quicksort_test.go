package test_test

import (
	"fmt"
	"testing"
)

func Test_sort(t *testing.T) {
	list := []int{12, 30, 5, 16, 5, 1, 20}

	bubble(list)

	fmt.Println(list)
}

func bubble(list []int) {
	for i := 0; i < len(list)-1; i++ {
		for i := 0; i < len(list)-1; i++ {
			if list[i] > list[i+1] {
				temp := list[i+1]
				list[i+1] = list[i]
				list[i] = temp
			}
		}
	}

}

func quick(list []int) {

}

//todo
func quickRecursion(list []int) {
	// 取一个基准值
	base := list[0]
	left := 0
	right := len(list) - 1
	for true {
		if left >= right {
			break
		}

		if list[left] >= base {
			//移动到右边 和right交换
			temp := list[left]
			list[right] = temp
			right--
		} else {
			left++
		}
	}

}
