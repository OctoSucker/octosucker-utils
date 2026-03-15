package graph

type Node struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Tool        string
	Next        []string
	Dynamic     bool
}
