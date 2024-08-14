package strings

import (
	"encoding/json"
	"log"
	"net/http"
	"stringy-go/internal/db"
)

type Server interface {
	Run() error
}

type server struct {
	db.DB
}

func NewServer(db db.DB) Server {
	return &server{
		db,
	}
}

func (s *server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/string", s.getString)
	mux.HandleFunc("/string", s.postString)
	return http.ListenAndServe(":6969", nil)
}

func (s *server) getString(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))

		return
	}

	active, err := s.DB.GetActive()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	_, _ = w.Write([]byte(active))

	w.WriteHeader(http.StatusOK)
}

func (s *server) postString(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("method not allowed"))

		return
	}

	var str string
	if err := json.NewDecoder(r.Body).Decode(&str); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := s.DB.Insert(str); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
