package forum 
import (
	"net/http"
	"log"
	s "forum/src/structs"
)
func RecoveryHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				ServerError(w)
			}
		}()
		next.ServeHTTP(w, r)
	}
}

func ServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	err := s.Tpl.ExecuteTemplate(w, "500.html", nil)
	if err != nil {
		log.Printf("Error rendering 500 page: %v", err)
		w.Write([]byte("500 Internal Server Error"))
	}
}

func NotFoundHandler(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	err := s.Tpl.ExecuteTemplate(w, "404.html", nil)
	if err != nil {
		log.Printf("Error rendering 404 page: %v", err)
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}