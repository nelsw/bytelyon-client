package model

type Page struct {
	Error       string         `json:"error"`
	URL         string         `json:"url"`
	Title       string         `json:"title"`
	Meta        map[string]any `json:"meta"`
	Links       []string       `json:"links"`
	Body        string         `json:"body"`
	ImgSrc      string         `json:"img_src"`
	ImgAlt      string         `json:"img_alt"`
	Description string         `json:"description"`
	Keywords    []string       `json:"keywords"`
	Screenshot  string         `json:"screenshot"`
}
