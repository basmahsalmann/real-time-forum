package forum

import (
	t "forum/src/database"
	s "forum/src/structs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)


func ValidateSession(w http.ResponseWriter, r *http.Request) (int, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return 0, false
	}

	sessionToken := cookie.Value
	var userID int
	err = t.Db.QueryRow("SELECT user_id FROM sessions WHERE token = ? AND expires_at > ?", sessionToken, time.Now()).Scan(&userID)
	if err != nil {
		// If session is invalid, clear the cookie and redirect to login
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now().Add(-1 * time.Hour),
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return 0, false
	}

	return userID, true
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var dbPassword string
		var userID int

		err := t.Db.QueryRow("SELECT id, password FROM user WHERE username = ?", username).Scan(&userID, &dbPassword)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data := struct {
				Error string
			}{
				Error: "Invalid username or password",
			}
			s.Tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			data := struct {
				Error string
			}{
				Error: "Invalid username or password",
			}
			s.Tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		sessionToken := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		// Delete any existing sessions for this user
		_, err = t.Db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
		if err != nil {
			log.Println("Login error: Failed to delete previous sessions:", err)
			http.Error(w, "Error managing session", http.StatusInternalServerError)
			return
		}

		// Create a new session
		_, err = t.Db.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)", userID, sessionToken, expiresAt)
		if err != nil {
			log.Println("Login error: Session creation failed:", err)
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: expiresAt,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := s.Tpl.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		log.Println("Template execution error:", err)
		ServerError(w)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		username := r.FormValue("username")
		password := r.FormValue("password")

		var errors []string

		// remove leading and trailing spaces to check if string is empty
		username = strings.TrimSpace(username)
		password = strings.TrimSpace(password)


		if email == "" || username == "" || password == "" {
			errors = append(errors, "All fields are required")
		}
		if len(email) < 15 || len(email) > 30 {
			errors = append(errors, "Email must be between 10 and 30 characters")
		}
		if len(username) < 3 || len(username) > 20 {
			errors = append(errors, "Username must be between 3 and 20 characters")
		}
		if len(password) < 8 || len(password) > 20 {
			errors = append(errors, "Password must be between 8 and 20 characters")
		}

		if len(errors) > 0 {
			data := struct {
				Errors   []string
				Email    string
				Username string
			}{
				Errors:   errors,
				Email:    email,
				Username: username,
			}
			w.WriteHeader(http.StatusBadRequest)
			s.Tpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		result, err := t.Db.Exec("INSERT INTO user (email, username, password) VALUES (?, ?, ?)", email, username, hashedPassword)
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: user.email" {
				errors = append(errors, "Email already exists")
			} else if err.Error() == "UNIQUE constraint failed: user.username" {
				errors = append(errors, "Username already exists")
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Println("Error inserting user:", err)
				return
			}
			data := struct {
				Errors   []string
				Email    string
				Username string
			}{
				Errors:   errors,
				Email:    email,
				Username: username,
			}
			w.WriteHeader(http.StatusBadRequest)
			s.Tpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		userID, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Error retrieving user ID", http.StatusInternalServerError)
			return
		}

		sessionToken := uuid.New().String()
		_, err = t.Db.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)", userID, sessionToken, time.Now().Add(24*time.Hour))
		if err != nil {
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.Tpl.ExecuteTemplate(w, "register.html", nil)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	sessionToken := cookie.Value
	_, err = t.Db.Exec("DELETE FROM sessions WHERE token = ?", sessionToken)
	if err != nil {
		http.Error(w, "Error logging out", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}