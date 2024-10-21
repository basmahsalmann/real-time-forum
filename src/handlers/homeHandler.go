package forum

import (
	"database/sql"
	"log"
	"net/http"
	t "real-time-forum/src/database"
	s "real-time-forum/src/structs"
	"strings"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var posts []s.Post
	var err error
	var category string
	var userID int
	var loggedIn bool

	// Check if the user is logged in
	cookie, err := r.Cookie("session_token")
	if err == nil {
		sessionToken := cookie.Value
		err = t.Db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", sessionToken).Scan(&userID)
		if err == nil {
			loggedIn = true
		}
	}

	// Get category from query parameters
	category = r.URL.Query().Get("category")

	// Initialize the base SQL query
	query := `
		SELECT 
			p.id, p.title, p.content, 
			 GROUP_CONCAT(DISTINCT c.name ORDER BY c.name ASC) AS categories, 
			u.username AS author
		FROM 
			post p
		JOIN 
			post_category pc ON p.id = pc.post_id
		JOIN
			category c ON pc.category_id = c.id
		JOIN 
			user u ON p.user_id = u.id`

	var args []interface{}
	conditions := []string{}

	// Apply category filter if selected
	if category != "" {
		conditions = append(conditions, "p.id IN (SELECT pc.post_id FROM post_category pc WHERE pc.category_id = ?)")
		args = append(args, category)
	}

	// Apply additional filters for logged-in users
	if loggedIn {
		filter := r.URL.Query().Get("filter")
		switch filter {
		case "my-posts":
			conditions = append(conditions, "u.id = ?")
			args = append(args, userID)
		case "liked-posts":
			conditions = append(conditions, "p.id IN (SELECT post_id FROM post_likes_dislikes WHERE user_id = ? AND type = 'like')")
			args = append(args, userID)
		}
	}

	// Combine the conditions into the query
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY p.id ORDER BY p.id DESC"

	// Execute the query with the conditions
	rows, err := t.Db.Query(query, args...)
	if err != nil {
		log.Println("Error fetching posts:", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Iterate over the result set and populate the posts map
	for rows.Next() {
		var post s.Post
		var categories string
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &categories, &post.Author)
		if err != nil {
			log.Println("Error scanning post:", err)
			http.Error(w, "Error processing posts", http.StatusInternalServerError)
			return
		}

		if len(post.Content) > 10 { 
			post.Content = post.Content[:10]+"..."
		}

		// Split the categories string into a slice and trim whitespace
		categorySlice := strings.Split(categories, ",")
		for i := range categorySlice {
			categorySlice[i] = strings.TrimSpace(categorySlice[i])
		}
		post.Categories = categorySlice

		// Get likes and dislikes for the post
		post.Likes, post.Dislikes = getPostLikesDislikes(t.Db, post.ID)

		// Append the post to the slice
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Println("Error after scanning rows:", err)
		http.Error(w, "Error processing posts", http.StatusInternalServerError)
		return
	}

	// Fetch all categories for the filter dropdown
	categories, err := getAllCategories(t.Db)
	if err != nil {
		log.Println("Error fetching categories:", err)
		categories = []s.Category{}
	}

	// Pass the posts, categories, and login status to the template
	data := struct {
		Posts      []s.Post
		LoggedIn   bool
		Categories []s.Category
	}{
		Posts:      posts,
		LoggedIn:   loggedIn,
		Categories: categories,
	}

	// Render the template
	err = s.Tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Println("Template execution error:", err)
	}

	if err != nil {
		ServerError(w)
		return
	}
}

func getAllCategories(db *sql.DB) ([]s.Category, error) {
	rows, err := db.Query("SELECT id, name FROM category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []s.Category
	for rows.Next() {
		var category s.Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}