package test

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// 动态规划 =  递归暴力 + 备忘录

// 不加备忘录的耗时为O(n^2) 指数级别耗时，时间复杂度爆炸
// 目标40
// 加了备忘录 33 微秒
// 不加备忘录  43334305 微妙
func Test_dp1(t *testing.T) {
	start := time.Now().UnixMicro()
	i := dp([]int{1, 2, 5}, 40)
	fmt.Println(i)
	end := time.Now().UnixMicro()
	fmt.Println("program time : ", end-start)
}

// 定义一个备忘录
// 这里数据结构 也可以用dp数组来表示 k->索引，v->索引对应的数值
//  dp数组 更节省空间
var memoMap = make(map[int]int)

// dp 函数：dp(n) 表示，输入一个目标金额 n，返回凑出目标金额 n 所需的最少硬币数量。
func dp(coins []int, amount int) int {
	if amount == 0 {
		return 0
	}
	if amount < 0 {
		return -1
	}

	res := math.MaxFloat64
	// 子问题，查询11-1 11-2 11-5 凑出的最少硬币数量 最终结果就是这个数量+1
	for _, coin := range coins {
		// 先查询备忘录，找不到在进行dp算法
		subRes, ok := memoMap[amount-coin]
		if !ok {
			subRes = dp(coins, amount-coin)
		}
		if subRes == -1 {
			continue
		}
		// 最终结果就是这个数量+1
		res = math.Min(res, float64(subRes+1))
	}
	if res == math.MaxFloat64 {
		return -1
	}
	memoMap[amount] = int(res)
	return int(res)
}
