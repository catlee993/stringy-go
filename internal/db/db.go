package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	Insert(string) error
	GetActive() (string, error)
	NextActive() error
	ActiveUpdater()
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
			select id from strings where active = true limit 1
		),
		next_active as (
			select coalesce(
				(select id from strings where id > (select id from current_active) limit 1),
				(select id from strings order by id asc limit 1)
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

func (s *source) ActiveUpdater() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.NextActive(); err != nil {
					log.Printf("Failed to set next active string: %v", err)
				} else {
					log.Println("Successfully set the next active string.")
				}
			}
		}
	}()
}

func Exists(addr *string) bool {
	a := DefaultDBAddress
	if addr != nil {
		a = *addr
	}
	if _, err := os.Stat(a); err != nil {
		return false
	}

	db, err := sql.Open("sqlite3", a)
	if err != nil {
		log.Fatal(err)
	}

	exists, eErr := tableExists(db, "strings")
	if eErr != nil {
		log.Fatal(eErr)
	}

	return exists
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
		active boolean not null default false
    );
    `

	db, err := sql.Open("sqlite3", a)
	if err != nil {
		log.Fatal(err)
	}

	if _, eErr := db.Exec(createTableSQL); eErr != nil {
		log.Fatal(eErr)
	}

	return nil
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

func tableExists(db *sql.DB, tableName string) (bool, error) {
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s';", tableName)
	var name string
	err := db.QueryRow(query).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return name == tableName, nil
}
