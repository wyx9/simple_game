package test

import (
	"fmt"
	"testing"
)

// 最大回文数
func longestPalindrome(s string) string {
	end := ""
	for i := 0; i < len(s); i++ {
		tempList := ""
		u := s[i]
		for j := i + 1; j < len(s); j++ {
			u1 := s[j]
			if u == u1 {
				tempList = s[i:j]
			}
		}
		if len(tempList) > len(end) {
			end = tempList
		}
	}

	return end
}

func Test1Code(t *testing.T) {
	//palindrome := longestPalindrome("asddasdas")
	//fmt.Println(palindrome)

	a := (8345 * 249) + (9728 * 249) + (9603 * 249) + (7149 * 217) + (8067 * 217) + (7070 * 217) + (5884 * 188) + (5452 * 188) + (4860 * 164) + (929 * 164)
	fmt.Println(a)
	fmt.Println(470000)
}
