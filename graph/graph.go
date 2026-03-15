package graph

const DefaultMaxNodes = 30
const MaxHistoryRepeat = 3

type Graph struct {
	Nodes    map[string]*Node
	Current  []string
	History  []string
	MaxNodes int
}

func NewGraph() *Graph {
	return &Graph{
		Nodes:    make(map[string]*Node),
		Current:  nil,
		History:  nil,
		MaxNodes: DefaultMaxNodes,
	}
}

func NewGraphFromNodes(nodes map[string]*Node) *Graph {
	if nodes == nil {
		nodes = make(map[string]*Node)
	}
	copy := make(map[string]*Node, len(nodes))
	for k, v := range nodes {
		copy[k] = v
	}
	return &Graph{
		Nodes:    copy,
		Current:  nil,
		History:  nil,
		MaxNodes: DefaultMaxNodes,
	}
}

func (g *Graph) AddNode(node *Node) {
	if node == nil {
		return
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node)
	}
	g.Nodes[node.Name] = node
}

func (g *Graph) AddEdge(from, to string) {
	n := g.Nodes[from]
	if n == nil {
		return
	}
	for _, s := range n.Next {
		if s == to {
			return
		}
	}
	n.Next = append(n.Next, to)
}

func (g *Graph) HasNode(name string) bool {
	_, ok := g.Nodes[name]
	return ok
}

func (g *Graph) GetNode(name string) *Node {
	return g.Nodes[name]
}

func (g *Graph) AddDynamicNode(node *Node) bool {
	if node == nil {
		return false
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]*Node)
	}
	if g.HasNode(node.Name) {
		return false
	}
	max := g.MaxNodes
	if max <= 0 {
		max = DefaultMaxNodes
	}
	if len(g.Nodes) >= max {
		return false
	}
	node.Dynamic = true
	g.Nodes[node.Name] = node
	g.Current = []string{node.Name}
	return true
}

func (g *Graph) HistoryRepeatedCount(name string) int {
	var n int
	for _, h := range g.History {
		if h == name {
			n++
		}
	}
	return n
}

func (g *Graph) HistoryRepeatedTooMuch() bool {
	if len(g.History) == 0 {
		return false
	}
	seen := make(map[string]int)
	for _, name := range g.History {
		seen[name]++
		if seen[name] > MaxHistoryRepeat {
			return true
		}
	}
	return false
}

func (g *Graph) CurrentNodes() []*Node {
	if len(g.Current) == 0 {
		return nil
	}
	out := make([]*Node, 0, len(g.Current))
	seen := make(map[string]bool)
	for _, name := range g.Current {
		if seen[name] {
			continue
		}
		seen[name] = true
		if n, ok := g.Nodes[name]; ok {
			out = append(out, n)
		}
	}
	return out
}

func (g *Graph) Advance(name string) {
	g.History = append(g.History, name)
	node := g.Nodes[name]
	if node == nil || len(node.Next) == 0 {
		g.Current = nil
		return
	}
	next := make([]string, 0, len(node.Next))
	seen := make(map[string]bool)
	for _, n := range node.Next {
		if !seen[n] && g.Nodes[n] != nil {
			seen[n] = true
			next = append(next, n)
		}
	}
	g.Current = next
}

func (g *Graph) SetCurrent(names []string) {
	g.Current = names
}
