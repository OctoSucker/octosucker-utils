package graph

type Node struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	Tool        string                 `json:"tool"`
	Next        []string               `json:"next"`
	Dynamic     bool                   `json:"dynamic"`
}
