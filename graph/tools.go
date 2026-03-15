package graph

func ToolDefsFromGraph(g *Graph) []map[string]interface{} {
	nodes := g.CurrentNodes()
	if len(nodes) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(nodes))
	for _, n := range nodes {
		fn := nodeToFunction(n)
		if fn != nil {
			out = append(out, fn)
		}
	}
	return out
}

func nodeToFunction(node *Node) map[string]interface{} {
	if node == nil {
		return nil
	}
	params := node.InputSchema
	if params == nil {
		params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
	}
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        node.Name,
			"description": node.Description,
			"parameters":  params,
		},
	}
}
