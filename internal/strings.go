package strings

import (
	"io"
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
	mux.HandleFunc("/save", s.postString)
	return http.ListenAndServe(":6969", mux)
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	str := string(body)

	if err := s.DB.Insert(str); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
