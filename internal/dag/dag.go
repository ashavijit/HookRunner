package dag

import (
	"github.com/ashavijit/hookrunner/internal/config"
)

type Node struct {
	Hook     config.Hook
	Children []*Node
	InDegree int
}

type Graph struct {
	Nodes map[string]*Node
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
	}
}

func BuildGraph(hooks []config.Hook) *Graph {
	g := NewGraph()

	for _, h := range hooks {
		g.Nodes[h.Name] = &Node{
			Hook:     h,
			Children: make([]*Node, 0),
			InDegree: 0,
		}
	}

	for _, h := range hooks {
		if h.After != "" {
			if parent, exists := g.Nodes[h.After]; exists {
				child := g.Nodes[h.Name]
				parent.Children = append(parent.Children, child)
				child.InDegree++
			}
		}
	}

	return g
}

func (g *Graph) TopologicalSort() [][]*Node {
	var levels [][]*Node
	inDegree := make(map[string]int)

	for name, node := range g.Nodes {
		inDegree[name] = node.InDegree
	}

	for {
		var currentLevel []*Node

		for name, degree := range inDegree {
			if degree == 0 {
				currentLevel = append(currentLevel, g.Nodes[name])
				inDegree[name] = -1
			}
		}

		if len(currentLevel) == 0 {
			break
		}

		for _, node := range currentLevel {
			for _, child := range node.Children {
				inDegree[child.Hook.Name]--
			}
		}

		levels = append(levels, currentLevel)
	}

	return levels
}

func (g *Graph) HasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(name string) bool
	hasCycle = func(name string) bool {
		visited[name] = true
		recStack[name] = true

		node := g.Nodes[name]
		for _, child := range node.Children {
			if !visited[child.Hook.Name] {
				if hasCycle(child.Hook.Name) {
					return true
				}
			} else if recStack[child.Hook.Name] {
				return true
			}
		}

		recStack[name] = false
		return false
	}

	for name := range g.Nodes {
		if !visited[name] {
			if hasCycle(name) {
				return true
			}
		}
	}

	return false
}

func (g *Graph) GetExecutionPlan() [][]config.Hook {
	levels := g.TopologicalSort()
	plan := make([][]config.Hook, len(levels))

	for i, level := range levels {
		plan[i] = make([]config.Hook, len(level))
		for j, node := range level {
			plan[i][j] = node.Hook
		}
	}

	return plan
}
