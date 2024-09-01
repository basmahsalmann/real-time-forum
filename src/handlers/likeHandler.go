package forum

import (
	"database/sql"
	t "forum/src/database"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if the like belongs to a comment or post
	isPost := strings.Contains(path, "post")
	isComment := strings.Contains(path, "comment")

	var commentID, postID int
	var err error

	if isPost {
		postIDStr := path[len("/like/post/"):]
		postID, err = strconv.Atoi(postIDStr)
		if err != nil || postID <= 0 {
			log.Println("Invalid post ID:", err)
			http.NotFound(w, r)
			return
		}
	} else {
		parts := strings.Split(path, "/")

		// The comment ID will be the last part of the URL
		commentIDStr := parts[len(parts)-1]

		// Convert the comment ID from string to int
		commentID, err = strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}
	}

	// Get the user id
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
		// http.Error(w, "Invalid credenxtials", http.StatusUnauthorized)
		return
	}

	var ID int
	var table, idColumn string

	if isComment {
		ID = commentID
		table = "comment_likes_dislikes"
		idColumn = "comment_id"
		log.Println("Comment ID:", commentID)
	} else {
		ID = postID
		table = "post_likes_dislikes"
		idColumn = "post_id"
	}

	var existingType string
	err = t.Db.QueryRow("SELECT type FROM "+table+" WHERE user_id = ? AND "+idColumn+" = ?", userID, ID).Scan(&existingType)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Error checking like status", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// User has not liked or disliked yet, so insert the like
		_, err = t.Db.Exec("INSERT INTO "+table+" (user_id, "+idColumn+", type) VALUES (?, ?, 'like')", userID, ID)
	} else if existingType == "dislike" {
		// User previously disliked, so update to like
		_, err = t.Db.Exec("UPDATE "+table+" SET type = 'like' WHERE user_id = ? AND "+idColumn+" = ?", userID, ID)
	} else {
		// User already liked, so remove the like
		log.Println(commentID)
		_, err = t.Db.Exec("DELETE FROM "+table+" WHERE user_id = ? AND "+idColumn+" = ? AND type = 'like'", userID, ID)
	}

	if err != nil {
		http.Error(w, "Error updating like", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if the like belongs to a comment or post
	isPost := strings.Contains(path, "post")
	isComment := strings.Contains(path, "comment")

	var commentID, postID int
	var err error

	if isPost {
		postIDStr := path[len("/dislike/post/"):]
		postID, err = strconv.Atoi(postIDStr)
		if err != nil || postID <= 0 {
			log.Println("Invalid post ID:", err)
			http.NotFound(w, r)
			return
		}
	} else {
		parts := strings.Split(path, "/")

		// The comment ID will be the last part of the URL
		commentIDStr := parts[len(parts)-1]

		// Convert the comment ID from string to int
		commentID, err = strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}
	}

	// Get the user id
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
		// http.Error(w, "Invalid credenxtials", http.StatusUnauthorized)
		return
	}

	var ID int
	var table, idColumn string

	if isComment {
		ID = commentID
		table = "comment_likes_dislikes"
		idColumn = "comment_id"
		log.Println("Comment ID:", commentID)
	} else {
		ID = postID
		table = "post_likes_dislikes"
		idColumn = "post_id"
	}

	var existingType string
	err = t.Db.QueryRow("SELECT type FROM "+table+" WHERE user_id = ? AND "+idColumn+"  = ?", userID, ID).Scan(&existingType)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Error checking like status", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// User has not liked or disliked yet, so insert the like
		_, err = t.Db.Exec("INSERT INTO "+table+" (user_id, "+idColumn+" , type) VALUES (?, ?, 'dislike')", userID, ID)
	} else if existingType == "like" {
		// User previously disliked, so update to like
		_, err = t.Db.Exec("UPDATE "+table+" SET type = 'dislike' WHERE user_id = ? AND "+idColumn+"  = ?", userID, ID)
	} else {
		// User already liked, so remove the like
		_, err = t.Db.Exec("DELETE FROM "+table+" WHERE user_id = ? AND "+idColumn+"  = ? AND type = 'dislike'", userID, ID)
	}

	if err != nil {
		http.Error(w, "Error updating like", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func getPostLikesDislikes(db *sql.DB, postID int) (likesCount int, dislikesCount int) {
	// Query to get the number of likes
	err := db.QueryRow("SELECT count(*) FROM post_likes_dislikes WHERE post_id = ? AND type = 'like'", postID).Scan(&likesCount)
	if err != nil {
		return 0, 0
	}

	// Query to get the number of dislikes
	err = db.QueryRow("SELECT count(*) FROM post_likes_dislikes WHERE post_id = ? AND type = 'dislike'", postID).Scan(&dislikesCount)
	if err != nil {
		return 0, 0
	}

	return likesCount, dislikesCount
}

func getCommentLikesDislikes(db *sql.DB, postID int) (likesCount int, dislikesCount int) {
	// Query to get the number of likes
	err := db.QueryRow("SELECT count(*) FROM comment_likes_dislikes WHERE comment_id = ? AND type = 'like'", postID).Scan(&likesCount)
	if err != nil {
		return 0, 0
	}

	// Query to get the number of dislikes
	err = db.QueryRow("SELECT count(*) FROM comment_likes_dislikes WHERE comment_id = ? AND type = 'dislike'", postID).Scan(&dislikesCount)
	if err != nil {
		return 0, 0
	}

	return likesCount, dislikesCount
}