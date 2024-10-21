package forum

import (
	"database/sql"
	"log"
	"net/http"
	t "real-time-forum/src/database"
	s "real-time-forum/src/structs"
	"strconv"
	"strings"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	_, valid := ValidateSession(w, r)
	if !valid {
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryIDs := r.Form["categories"]

		var errors []string

		if len(title) < 5 || len(title) > 100 {
			errors = append(errors, "Title must be between 5 and 100 characters")
		}
		if len(content) < 10 || len(content) > 500 {
			errors = append(errors, "Content must be between 20 and 5000 characters")
		}
		if len(categoryIDs) == 0 {
			errors = append(errors, "Please select at least one category")
		}

		if len(errors) > 0 {
			categories := getCategories()
			data := struct {
				Errors     []string
				Categories []struct {
					ID   int
					Name string
				}
				Title   string
				Content string
			}{
				Errors:     errors,
				Categories: categories,
				Title:      title,
				Content:    content,
			}
			w.WriteHeader(http.StatusBadRequest)
			s.Tpl.ExecuteTemplate(w, "new-post.html", data)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var userID int
		sessionToken := cookie.Value
		err = t.Db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", sessionToken).Scan(&userID)
		if err != nil {
			log.Println(err)
			return
		}

		insertPostSQL := `INSERT INTO post (user_id, title, content) VALUES (?, ?, ?)`
		statement, err := t.Db.Prepare(insertPostSQL)
		if err != nil {
			log.Fatalln(err.Error())
		}
		result, err := statement.Exec(userID, title, content)
		if err != nil {
			log.Println("Error inserting post:", err)
			http.Error(w, "Error creating post", http.StatusInternalServerError)
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			log.Println("Error retrieving post ID:", err)
			http.Error(w, "Error creating post", http.StatusInternalServerError)
			return
		}

		insertPostCategorySQL := `INSERT INTO post_category (post_id, category_id) VALUES (?, ?)`
		for _, categoryID := range categoryIDs {
			log.Println(categoryID)
			_, err := t.Db.Exec(insertPostCategorySQL, postID, categoryID)
			if err != nil {
				log.Println("Error inserting post_category:", err)
				http.Error(w, "Error associating post with category", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	categories := getCategories()
	data := struct {
		Errors     []string
		Categories []struct {
			ID   int
			Name string
		}
		Title   string
		Content string
	}{
		Categories: categories,
		Title:      "",
		Content:    "",
	}
	s.Tpl.ExecuteTemplate(w, "new-post.html", data)
}

func getCategories() []struct {
	ID   int
	Name string
} {
	categories := []struct {
		ID   int
		Name string
	}{}
	categoryrows, err := t.Db.Query("SELECT id, name FROM category;")
	if err != nil {
		log.Fatal(err)
	}
	defer categoryrows.Close()

	for categoryrows.Next() {
		var id int
		var name string
		err := categoryrows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		categories = append(categories, struct {
			ID   int
			Name string
		}{ID: id, Name: name})
	}

	if err = categoryrows.Err(); err != nil {
		log.Fatal(err)
	}

	return categories
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.URL.Path[len("/post/"):]
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		NotFoundHandler(w)
		return
	}

	var post s.Post
	var categoriesStr string
	err = t.Db.QueryRow(`
    SELECT
        p.id, p.title, p.content,
        GROUP_CONCAT(c.name) AS categories,
        u.username AS author
    FROM
        post p
    JOIN
        post_category pc ON p.id = pc.post_id
    JOIN
        category c ON pc.category_id = c.id
    JOIN
        user u ON p.user_id = u.id
    WHERE
        p.id = ?
    GROUP BY p.id
    `, postID).Scan(&post.ID, &post.Title, &post.Content, &categoriesStr, &post.Author)

	if err != nil {
		if err == sql.ErrNoRows {
			NotFoundHandler(w)
			return
		}
		log.Println("Error fetching post:", err)
		http.Error(w, "Error fetching post", http.StatusInternalServerError)
		return
	}

	post.Categories = strings.Split(categoriesStr, ",")

	comments, err := fetchCommentsWithLikesDislikes(t.Db, post.ID)
	if err != nil {
		log.Println("Error fetching comments with likes/dislikes:", err)
		comments = []s.Comment{}
	}

	post.Likes, post.Dislikes = getPostLikesDislikes(t.Db, postID)

	loggedIn := false
	_, err = r.Cookie("session_token")
	if err == nil {
		loggedIn = true
	}

	data := struct {
		Post     s.Post
		Comments []s.Comment
		LoggedIn bool
		Errors   []string
	}{
		Post:     post,
		Comments: comments,
		LoggedIn: loggedIn,
		Errors:   []string{},
	}

	err = s.Tpl.ExecuteTemplate(w, "post.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
