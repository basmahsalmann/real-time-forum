package forum

import (
	"log"
	"net/http"
	t "real-time-forum/src/database"
	s "real-time-forum/src/structs"
)

func CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch categories with their IDs
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

	// Pass the categories to the template
	data := struct {
		Categories []struct {
			ID   int
			Name string
		}
		LoggedIn bool
	}{
		Categories: categories,
		LoggedIn:   loggedIn,
	}

	err = s.Tpl.ExecuteTemplate(w, "categories.html", data)
	if err != nil {
		ServerError(w)
	}
}

func AddCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value
	var userID int
	err = t.Db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", sessionToken).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid session token", http.StatusUnauthorized)
		return
	}

	categoryName := r.FormValue("category")
	if categoryName == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	_, err = t.Db.Exec("INSERT INTO category (name, user_id) VALUES (?, ?)", categoryName, userID)
	if err != nil {
		http.Error(w, "Error adding category", http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
