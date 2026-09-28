package search

type Page struct {
	Img  string         `json:"screenshot"`
	Src  string         `json:"content"`
	Data map[string]any `json:"data"`
}
