package forum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	t "real-time-forum/src/database"
	s "real-time-forum/src/structs"
	"strconv"
	"strings"
)

func commentHandler(w http.ResponseWriter, r *http.Request) {
	_, valid := ValidateSession(w, r)
	if !valid {
		return
	}
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if len(content) > 500 {
		http.Error(w, "Comment content too long", http.StatusBadRequest)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 || parts[3] != "comment" {
		http.NotFound(w, r)
		return
	}

	postID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	content = r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Get the user ID from the session
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var userID int
	err = t.Db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Insert the comment into the database
	_, err = t.Db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	if err != nil {
		http.Error(w, "Error submitting comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page
	http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
}

func PostAndCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/comment") {
		commentHandler(w, r)
	} else {
		postHandler(w, r)
	}
}

func fetchCommentsWithLikesDislikes(db *sql.DB, postID int) ([]s.Comment, error) {
	// Query to fetch comments for a post
	rows, err := db.Query("SELECT comment_id, content, u.username as Auther FROM comments JOIN user u ON comments.user_id = u.id WHERE post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []s.Comment
	for rows.Next() {
		var comment s.Comment
		// Scan basic comment data
		err := rows.Scan(&comment.ID, &comment.Content, &comment.Author)
		if err != nil {
			log.Println("Error scanning comment:", err)
			continue
		}

		// Fetch likes and dislikes for each comment
		likesCount, dislikesCount := getCommentLikesDislikes(db, comment.ID)
		comment.Likes = likesCount
		comment.Dislikes = dislikesCount

		comments = append(comments, comment)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}