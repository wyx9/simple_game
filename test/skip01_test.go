package test_test

// 跳表

// 最大层级
const MAXLEVEL = 7

type Node struct {
	// 存储值
	Value int
	// 同一层前节点
	Prev *Node
	// 同一层后节点
	Next *Node
	// 下层同节点??
	Down *Node
}

type SkipList struct {
	// 层级
	Level int
	// 节点
	HeadNodeArr []*Node
}

// 查找节点
func (list SkipList) HasNode(value int) *Node {
	if list.Level >= 0 {
		// 只有层级在大于等于 0 的时候在进行循环判断，如果层级小于 0 说明是没有任何数据
		level := list.Level
		node := list.HeadNodeArr[level].Next
		for node != nil {
			if node.Value == value {
				// 如果节点的值 等于 传入的值 就说明包含这个节点 返回 true
				return node
			} else if node.Value > value {
				// 如果节点的值大于传入的值，就应该返回上个节点并进入下一层
				if node.Prev.Down == nil {
					if level-1 >= 0 {
						node = list.HeadNodeArr[level-1].Next
					} else {
						node = nil
					}
				} else {
					node = node.Prev.Down
				}
				level -= 1
			} else if node.Value < value {
				// 如果节点的值小于传入的值就进入下一个节点，如果下一个节点是 nil，说明本层已经查完了，进入下一层，且从下一层的头部开始
				node = node.Next
				if node == nil {
					level -= 1
					if level >= 0 {
						// 如果不是最底层继续进入到下一层
						node = list.HeadNodeArr[level].Next
					}
				}
			}
		}
	}
	return nil
}
