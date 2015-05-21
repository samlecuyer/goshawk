package fs

import (
	"path"
	"strings"
)

type Post struct {
	// this is the "filename"
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Raw       string `json:"raw"`
	Published bool   `json:"published"`
}

func FromMap(name string, m map[string]interface{}) *Post {
	p := new(Post)
	title, ok := m["title"].(string)
	if ok {
		p.Title = title
	} else {
		p.Title = "Untitled"
	}
	slug, ok := m["slug"].(string)
	if ok {
		p.Slug = slug
	} else {
		if name == "/posts/new" {
			lc := strings.ToLower(p.Title)
			lcs := strings.Split(lc, " ")
			lc = strings.Join(lcs, "-")
			p.Slug = lc
		} else {
			_, name := path.Split(name)
			p.Slug = name
		}
	}
	return p
}
