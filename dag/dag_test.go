package dag

import "testing"

func TestAddIdenticalVertex(t *testing.T) {
	dag := New()
	if err := dag.AddVertex("A", 1); err != nil {
		t.Fatal("expected no error")
	}
	if err := dag.AddVertex("A", 1); err == nil {
		t.Fatal("expected to return an error: vertex already in DAG")
	}
}

func TestAddIdenticalEdge(t *testing.T) {
	dag := New()
	if err := dag.AddVertex("A", 1); err != nil {
		t.Fatal("expected no error")
	}
	if err := dag.AddVertex("B", 1); err != nil {
		t.Fatal("expected no error")
	}
	if err := dag.AddEdge("B", "A"); err != nil {
		t.Fatalf("expected to not return an error but got %s", err)
	}
	if err := dag.AddEdge("B", "A"); err == nil {
		t.Fatalf("expected to return an error but")
	}
}

func TestAddEdge(t *testing.T) {
	dag := New()
	if err := dag.AddVertex("A", 1); err != nil {
		t.Fatal("expected no error")
	}
	if err := dag.AddVertex("B", 1); err != nil {
		t.Fatal("expected no error")
	}
	if err := dag.AddEdge("B", "A"); err != nil {
		t.Fatalf("expected to not return an error but got %s", err)
	}
	a, err := dag.GetVertex("A")
	if err != nil {
		t.Fatal("expected no error")
	}
	b, err := dag.GetVertex("B")
	if err != nil {
		t.Fatal("expected no error")
	}
	if !a.HasChild(b) {
		t.Fatal("expected a to have child b")
	}
	if !b.HasParent(a) {
		t.Fatal("expected b to have parent a")
	}
}
