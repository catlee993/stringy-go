package main

import (
	"fmt"
	strings "stringy-go/internal"
	"stringy-go/internal/db"
)

func main() {
	// nil will default to default address
	if db.Exists(nil) == false {
		err := db.Create(nil)
		if err != nil {
			panic(err)
		}
	}
	dbase := db.New(nil)
	s := strings.NewServer(dbase)

	nErr := dbase.NextActive()
	if nErr != nil {
		fmt.Println(nErr)
	}

	dbase.ActiveUpdater()

	err := s.Run()
	if err != nil {
		fmt.Println(err)
	}

	_ = db.Close(dbase)
}
