package forum

import (
	"html/template"
	"strings"
	"log"
	h "forum/src/structs"
)

func init() {
	InitDB()

	funcMap := template.FuncMap{
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br>", -1))
		},
	}

	var err error
	h.Tpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}