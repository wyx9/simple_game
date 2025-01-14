package test_test

import (
	"container/list"
	"fmt"
	"testing"
)

// 双向链表
func PrintList(list *list.List) {
	fmt.Println("this list:")
	for e := list.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
	}

}

func Test1(t *testing.T) {
	// 双向链表
	l := list.New()

	// 向尾部插入
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)
	PrintList(l)

	//输出首部元素的值
	fmt.Println("start", l.Front().Value)
	//输出尾部元素的值
	fmt.Println("end", l.Back().Value)
	// 向首部插入
	l.PushFront(0)
	PrintList(l)

}
