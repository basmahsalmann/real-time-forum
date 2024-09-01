package main

import (
	"fmt"
	h "forum/src/handlers"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			h.NotFoundHandler(w)
			return
		}
		h.RecoveryHandler(h.HomeHandler)(w, r)
	})
	http.HandleFunc("/login", h.RecoveryHandler(h.LoginHandler))
	http.HandleFunc("/register", h.RecoveryHandler(h.RegisterHandler))
	http.HandleFunc("/logout", h.RecoveryHandler(h.LogoutHandler))
	http.HandleFunc("/create-post", h.RecoveryHandler(h.CreatePostHandler))
	http.HandleFunc("/post/", h.RecoveryHandler(h.PostAndCommentHandler))
	http.HandleFunc("/like/", h.RecoveryHandler(h.LikeHandler))
	http.HandleFunc("/dislike/", h.RecoveryHandler(h.DislikeHandler))
	http.HandleFunc("/new-post", h.RecoveryHandler(h.CreatePostHandler))
	http.HandleFunc("/categories", h.RecoveryHandler(h.CategoriesHandler))
	http.HandleFunc("/add-category", h.RecoveryHandler(h.AddCategoryHandler))

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	message := fmt.Sprintf(" Server started at http://localhost:%s/\n", port)
	log.Print(message)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

