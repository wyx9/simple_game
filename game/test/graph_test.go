package test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

// 定义节点结构体
type Node struct {
	ID       int         // 节点 ID
	Value    interface{} // 节点值
	Adjacent []*Node     // 相邻节点列表
}

// 定义图结构体
type Graph struct {
	Nodes []*Node // 节点列表
}

// 添加节点到图中
func (g *Graph) AddNode(id int, value interface{}) *Node {
	node := &Node{
		ID:       id,
		Value:    value,
		Adjacent: []*Node{},
	}
	g.Nodes = append(g.Nodes, node)
	return node
}

// 添加边到节点中
func (n *Node) AddEdge(node *Node) {
	n.Adjacent = append(n.Adjacent, node)
}

// 根据节点 ID 查找节点
func (g *Graph) FindNodeByID(id int) *Node {
	for _, node := range g.Nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

// 深度优先搜索遍历图
func DFS(node *Node, visited map[int]bool) {
	if visited[node.ID] {
		return
	}
	fmt.Printf("%v -> ", node.Value)
	visited[node.ID] = true
	for _, adj := range node.Adjacent {
		DFS(adj, visited)
	}
}

// 广度优先搜索遍历图
func BFS(start *Node) {
	queue := make([]*Node, 0)
	visited := make(map[int]bool)

	queue = append(queue, start)
	visited[start.ID] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		fmt.Printf("%v -> ", node.Value)
		for _, adj := range node.Adjacent {
			if !visited[adj.ID] {
				visited[adj.ID] = true
				queue = append(queue, adj)
			}
		}
	}
}

// Dijkstra 最短路径算法
func Dijkstra(g *Graph, start *Node, end *Node) []int {
	dist := make(map[int]int) // 距离表
	prev := make(map[int]int) // 前驱表
	unvisited := make(map[int]bool)

	// 初始化
	for _, node := range g.Nodes {
		dist[node.ID] = math.MaxInt32
		unvisited[node.ID] = true
	}
	dist[start.ID] = 0

	// 遍历所有节点
	for len(unvisited) > 0 {
		var current *Node
		minDist := math.MaxInt32

		// 获取距离起点最近的节点
		for _, node := range g.Nodes {
			if dist[node.ID] < minDist && unvisited[node.ID] {
				minDist = dist[node.ID]
				current = node
			}
		}

		// 遍历该节点的邻居节点，更新距离和前驱
		for _, adj := range current.Adjacent {
			if unvisited[adj.ID] {
				alt := dist[current.ID] + 1 // 计算新的距离值
				if alt < dist[adj.ID] {
					dist[adj.ID] = alt
					prev[adj.ID] = current.ID
				}
			}
		}

		delete(unvisited, current.ID)
	}

	// 构建路径
	path := []int{}
	node := end.ID
	for node != start.ID {
		path = append([]int{node}, path...)
		node = prev[node]
	}
	path = append([]int{start.ID}, path...)

	return path
}
func PrintGraph(g *Graph) {
	for _, node := range g.Nodes {
		fmt.Printf("Node %v -> ", node.Value)
		for _, adj := range node.Adjacent {
			fmt.Printf("%v ", adj.Value)
		}
		fmt.Println()
	}
}

func GenerateGraph(n int, p float64) *Graph {
	g := &Graph{Nodes: []*Node{}}

	// 添加节点
	for i := 1; i <= n; i++ {
		g.AddNode(i, i)
	}

	// 添加边
	for i := 1; i <= n; i++ {
		for j := i + 1; j <= n; j++ {
			if rand.Float64() < p {
				nodeI := g.FindNodeByID(i)
				nodeJ := g.FindNodeByID(j)
				nodeI.AddEdge(nodeJ)
				nodeJ.AddEdge(nodeI)
			}
		}
	}

	return g
}

func Test_graph1(t *testing.T) {
	//graph := GenerateGraph(10, 5000)
	//PrintGraph(graph)
	rand.Seed(time.Now().Unix())
	fmt.Println(rand.Float64() * 100)
	//for i := 0; i < 10; i++ {
	//	fmt.Println(rand.Float64())
	//}
	//for i := 0; i < 10; i++ {
	//	fmt.Println(rand.Int())
	//}
}
