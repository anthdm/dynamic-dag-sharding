package dag

import (
	"fmt"
	"os"

	"github.com/funkygao/golib/str"
)

// Dag represents a "Directed Acyclic Graph".
type Dag struct {
	vertices map[string]*Vertex
}

// New returns new Dag object.
func New() *Dag {
	return &Dag{
		vertices: make(map[string]*Vertex),
	}
}

// AddVertex adds a new Vertex to the DAG, given a name and value. This will
// return an error if the vertex allready is in the DAG.
func (d *Dag) AddVertex(name string, v interface{}) error {
	if d.HasVertex(name) {
		return fmt.Errorf("cannot add the same vertex (%s) more then once", name)
	}
	d.vertices[name] = &Vertex{
		name:     name,
		value:    v,
		parents:  []*Vertex{},
		children: []*Vertex{},
	}
	return nil
}

// AddEdges creates multiple edges from the given child to the given parents.
func (d *Dag) AddEdges(child string, parents ...string) error {
	for _, parent := range parents {
		if err := d.AddEdge(child, parent); err != nil {
			return err
		}
	}
	return nil
}

// AddEdge creates an edge in the direction from child -> parent.
func (d *Dag) AddEdge(child, parent string) error {
	from, err := d.GetVertex(child)
	if err != nil {
		return err
	}
	to, err := d.GetVertex(parent)
	if err != nil {
		return err
	}
	if err := to.addChild(from); err != nil {
		return err
	}
	return from.addParent(to)
}

// GetVertex returns the Vertex found by the given name.
func (d *Dag) GetVertex(name string) (*Vertex, error) {
	vertex, ok := d.vertices[name]
	if !ok {
		return nil, fmt.Errorf("could not find vertex with name %s", name)
	}
	return vertex, nil
}

// HasVertex returns true whether the DAG contains the given name of the vertex.
func (d *Dag) HasVertex(name string) bool {
	_, ok := d.vertices[name]
	return ok
}

func (d *Dag) CreateGraph(path string) string {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	builder := str.NewStringBuilder()
	builder.WriteString("digraph depgraph {\n\trankdir=LR;\n")
	for _, vertex := range d.vertices {
		vertex.createGraph(builder)
	}
	builder.WriteString("}\n")
	file.WriteString(builder.String())
	return builder.String()
}

// Vertex represents a vertex in the graph where its underlying value is of type
// interface. Lib users are repsonsible converting it to their appropriate type.
type Vertex struct {
	name     string
	value    interface{}
	parents  []*Vertex
	children []*Vertex
	indegree uint
}

func (v *Vertex) addChild(child *Vertex) error {
	if v.HasChild(child) {
		return fmt.Errorf("vertex already contains child %v", child)
	}
	v.children = append(v.children, child)
	return nil
}

// HashChild returns true whether the Vertex has the given child.
func (v *Vertex) HasChild(child *Vertex) bool {
	for _, val := range v.children {
		if child == val {
			return true
		}
	}
	return false
}

// HashChild returns true whether the Vertex has the given parent.
func (v *Vertex) HasParent(parent *Vertex) bool {
	for _, val := range v.parents {
		if parent == val {
			return true
		}
	}
	return false
}

// Parents returns the partent vertices of v.
func (v *Vertex) Parents() []*Vertex {
	return v.parents
}

// Children returns the children vertices of v.
func (v *Vertex) Children() []*Vertex {
	return v.children
}

func (v *Vertex) addParent(parent *Vertex) error {
	if v.HasParent(parent) {
		return fmt.Errorf("vertex already contains child %v", parent)
	}
	v.parents = append(v.parents, parent)
	return nil
}

func (v *Vertex) createGraph(builder *str.StringBuilder) {
	if len(v.parents) == 0 {
		builder.WriteString(fmt.Sprintf("\t\"%s\";\n", v.name))
		return
	}
	for _, parent := range v.parents {
		builder.WriteString(fmt.Sprintf(`%s -> %s [label="%v"]`, v.name, parent.name, v.value))
		builder.WriteString("\r\n")
	}
}
