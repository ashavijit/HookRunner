package dag

import (
	"testing"

	"github.com/ashavijit/hookrunner/internal/config"
)

func TestBuildGraph(t *testing.T) {
	hooks := []config.Hook{
		{Name: "format"},
		{Name: "lint", After: "format"},
		{Name: "test", After: "lint"},
	}

	g := BuildGraph(hooks)

	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}

	if g.Nodes["format"].InDegree != 0 {
		t.Error("format should have in-degree 0")
	}

	if g.Nodes["lint"].InDegree != 1 {
		t.Error("lint should have in-degree 1")
	}

	if g.Nodes["test"].InDegree != 1 {
		t.Error("test should have in-degree 1")
	}
}

func TestTopologicalSort(t *testing.T) {
	hooks := []config.Hook{
		{Name: "format"},
		{Name: "lint", After: "format"},
		{Name: "test", After: "lint"},
	}

	g := BuildGraph(hooks)
	levels := g.TopologicalSort()

	if len(levels) != 3 {
		t.Errorf("expected 3 levels, got %d", len(levels))
	}

	if levels[0][0].Hook.Name != "format" {
		t.Error("first level should contain format")
	}

	if levels[1][0].Hook.Name != "lint" {
		t.Error("second level should contain lint")
	}

	if levels[2][0].Hook.Name != "test" {
		t.Error("third level should contain test")
	}
}

func TestTopologicalSort_Parallel(t *testing.T) {
	hooks := []config.Hook{
		{Name: "format"},
		{Name: "lint"},
		{Name: "security"},
		{Name: "test", After: "lint"},
	}

	g := BuildGraph(hooks)
	levels := g.TopologicalSort()

	if len(levels) != 2 {
		t.Errorf("expected 2 levels, got %d", len(levels))
	}

	if len(levels[0]) != 3 {
		t.Errorf("first level should have 3 parallel hooks, got %d", len(levels[0]))
	}

	if len(levels[1]) != 1 {
		t.Errorf("second level should have 1 hook, got %d", len(levels[1]))
	}
}

func TestHasCycle_NoCycle(t *testing.T) {
	hooks := []config.Hook{
		{Name: "format"},
		{Name: "lint", After: "format"},
		{Name: "test", After: "lint"},
	}

	g := BuildGraph(hooks)

	if g.HasCycle() {
		t.Error("expected no cycle")
	}
}

func TestHasCycle_HasCycle(t *testing.T) {
	g := NewGraph()

	g.Nodes["a"] = &Node{Hook: config.Hook{Name: "a"}}
	g.Nodes["b"] = &Node{Hook: config.Hook{Name: "b"}}
	g.Nodes["c"] = &Node{Hook: config.Hook{Name: "c"}}

	g.Nodes["a"].Children = []*Node{g.Nodes["b"]}
	g.Nodes["b"].Children = []*Node{g.Nodes["c"]}
	g.Nodes["c"].Children = []*Node{g.Nodes["a"]}

	if !g.HasCycle() {
		t.Error("expected cycle to be detected")
	}
}

func TestGetExecutionPlan(t *testing.T) {
	hooks := []config.Hook{
		{Name: "format"},
		{Name: "lint", After: "format"},
		{Name: "test", After: "lint"},
	}

	g := BuildGraph(hooks)
	plan := g.GetExecutionPlan()

	if len(plan) != 3 {
		t.Errorf("expected 3 levels in plan, got %d", len(plan))
	}

	if plan[0][0].Name != "format" {
		t.Error("first hook should be format")
	}

	if plan[1][0].Name != "lint" {
		t.Error("second hook should be lint")
	}

	if plan[2][0].Name != "test" {
		t.Error("third hook should be test")
	}
}

func TestGetExecutionPlan_Complex(t *testing.T) {
	hooks := []config.Hook{
		{Name: "a"},
		{Name: "b"},
		{Name: "c", After: "a"},
		{Name: "d", After: "a"},
		{Name: "e", After: "c"},
	}

	g := BuildGraph(hooks)
	plan := g.GetExecutionPlan()

	if len(plan) < 2 {
		t.Error("expected at least 2 levels")
	}

	level0Names := make(map[string]bool)
	for _, h := range plan[0] {
		level0Names[h.Name] = true
	}

	if !level0Names["a"] || !level0Names["b"] {
		t.Error("first level should contain a and b")
	}
}

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g.Nodes == nil {
		t.Error("Nodes should be initialized")
	}

	if len(g.Nodes) != 0 {
		t.Error("Nodes should be empty")
	}
}
