package db

import (
	"database/sql"
	"log"
	"os"
)

type DB interface {
	Insert(string) error
	GetActive() (string, error)
	NextActive() error
}

type source struct {
	db   *sql.DB
	addr string
}

const DefaultDBAddress = "./strings.db"
const driver = "sqlite3"

func New(addr *string) DB {
	a := DefaultDBAddress
	if addr != nil {
		a = *addr
	}

	db, err := Open(a)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (s *source) Insert(str string) error {
	_, err := s.db.Exec("insert into strings (string) values (?)", str)

	return err
}

func (s *source) NextActive() error {
	_, err := s.db.Exec(`
		with current_active as (
			select id from strings where active = true
		),
		next_active as (
			select coalesce(
				(select id from strings where id = (select id from current_active) + 1),
				1
			) as id
		)
		update strings
		set active = case
			when id = (select id from next_active) then true
			else false
		end;
	`)

	return err
}

func (s *source) GetActive() (string, error) {
	var str string
	err := s.db.QueryRow(`
		select string from strings 
	  	where active = true;
	`).Scan(&str)

	return str, err
}

func Exists(addr *string) bool {
	a := DefaultDBAddress
	if addr != nil {
		a = *addr
	}
	if _, err := os.Stat(a); err != nil {
		return false
	}

	return true
}

func Create(addr *string) error {
	a := DefaultDBAddress
	if addr != nil {
		a = *addr
	}

	_, err := os.Create(a)

	// Create a table for storing quotes
	createTableSQL := `create table if not exists strings (
		id integer not null primary key,
		string text not null unique,
		active boolean not null
    );
    `

	db, err := sql.Open("sqlite3", a)

	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	return err
}

func Open(addr string) (DB, error) {
	db, err := sql.Open(driver, addr)
	if err != nil {
		log.Fatal(err)
	}

	return &source{db: db}, nil
}

func Close(d DB) error {
	return d.(*source).db.Close()
}
