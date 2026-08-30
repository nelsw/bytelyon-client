package model

type Page struct {
	Domain     string              `json:"domain"`
	Meta       map[string][]string `json:"meta"`
	Screenshot []byte              `json:"screenshot"`
	Title      string              `json:"title"`
	URL        string              `json:"url"`
	Index      int                 `json:"index"`
	Kind       string              `json:"kind"`
}
