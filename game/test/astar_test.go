package test

import (
	"container/heap"
	"fmt"
	"testing"
)

const (
	rows = 7
	cols = 7
)

// 地图类型定义
type Map [rows][cols]int

// 节点类型定义
type node struct {
	row, col int // 节点坐标
	f, g, h  int // f(n), g(n), h(n)值
	parent   *node
}

// 节点切片类型定义，用于实现堆接口
type nodes []*node

func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Less(i, j int) bool { return ns[i].f < ns[j].f }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }

func (ns *nodes) Push(x interface{}) {
	node := x.(*node)
	*ns = append(*ns, node)
}

func (ns *nodes) Pop() interface{} {
	old := *ns
	n := len(old)
	node := old[n-1]
	*ns = old[:n-1]
	return node
}

// A*算法主体函数
func AStar(m Map, start, end [2]int) []*node {
	// 创建起点和终点节点
	startnode := &node{start[0], start[1], 0, 0, 0, nil}
	endnode := &node{end[0], end[1], 0, 0, 0, nil}

	// 初始化开放列表和关闭列表
	openList := make(nodes, 0)
	closeList := make(map[int]*node)

	// 将起点加入开放列表 压栈
	heap.Push(&openList, startnode)

	// 不断遍历开放列表直到找到终点或者列表为空
	for len(openList) > 0 {
		// 取出f(n)值最小的节点
		curnode := heap.Pop(&openList).(*node)

		// 如果该节点是终点，则返回路径
		if curnode.row == endnode.row && curnode.col == endnode.col {
			path := make([]*node, 0)
			for curnode != nil {
				path = append(path, curnode)
				curnode = curnode.parent
			}
			return path
		}

		// 将该节点加入关闭列表
		closeList[curnode.row*cols+curnode.col] = curnode

		// 遍历该节点的邻居节点
		for _, dir := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			// 计算邻居节点坐标
			row, col := curnode.row+dir[0], curnode.col+dir[1]

			// 判断邻居节点是否在地图范围内，并且不是障碍物
			if row >= 0 && row < rows && col >= 0 && col < cols && m[row][col] == 0 {
				// 计算g(n)和h(n)值
				// g 距离代价距离起点的距离代价每走1步+1
				g := curnode.g + 1
				// h 曼哈顿距离它表示两点在网格状地图上沿着水平和垂直方向移动的最小步数之和
				h := abs(row-endnode.row) + abs(col-endnode.col)

				// 判断该节点是否已经在关闭列表中
				if _, ok := closeList[row*cols+col]; ok {
					continue
				}

				// 查找开放列表中是否已经存在该节点
				var childnode *node
				for _, node := range openList {
					if node.row == row && node.col == col {
						childnode = node
						break
					}
				}

				// 如果不存在，则创建新节点并将其加入开放列表
				if childnode == nil {
					childnode = &node{row, col, 0, g, h, curnode}
					heap.Push(&openList, childnode)
				} else if g < childnode.g {
					// 如果存在但是g(n)更小，则更新子节点的父节点和g(n)值
					childnode.g = g
					childnode.parent = curnode
				}
			}
		}
	}
	return nil
}

// 计算绝对值函数
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Test_a1(t *testing.T) {
	//1 代表阻碍
	var m Map = [rows][cols]int{
		{0, 1, 0, 0, 0, 0, 0},
		{0, 1, 0, 1, 1, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 1, 1},
		{0, 0, 0, 1, 1, 0, 0},
	}
	for i := len(m) - 1; i >= 0; i-- {
		fmt.Println(m[i])
	}
	// 执行A*算法，获取路径
	path := AStar(m, [2]int{0, 0}, [2]int{6, 6})

	// 不适用于走斜线的情况 2代表最短路径
	for i := len(path) - 1; i >= 0; i-- {
		row := path[i].row
		col := path[i].col
		m[row][col] = 2
		fmt.Printf("(%d,%d) ", path[i].row, path[i].col)
	}
	fmt.Print("\n\n")
	for i := len(m) - 1; i >= 0; i-- {
		fmt.Println(m[i])
	}

}
