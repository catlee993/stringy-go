package main

import (
	strings "stringy-go/internal"
	"stringy-go/internal/db"
)

func main() {
	dbase := db.New(nil)
	s := strings.NewServer(dbase)

	err := s.Run()
	if err != nil {
		panic(err)
	}
}
