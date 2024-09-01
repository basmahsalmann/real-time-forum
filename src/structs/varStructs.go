
package forum

import "html/template"

var Tpl *template.Template

type Post struct {
	ID               int
	Title            string
	Content          string
	Categories       []string
	Author           string
	Likes            int
	Dislikes         int
}

type Comment struct {
	ID       int
	Content  string
	Author   string
	Likes    int
	Dislikes int
}

type Category struct {
	ID   int
	Name string
}