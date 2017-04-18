package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func addTables() {
	db, err := sql.Open("sqlite3", "./local.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table adds (id integer not null primary key, val varchar(64) not null );
	create table roots (id integer not null primary key, val varchar(64) not null );
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func main() {
	http.HandleFunc("/add", addReq)
	http.HandleFunc("/remove", resetReq)
	http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil)
}

// PUT /add
func addReq(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Add Method")
}

// POST /reset
func resetReq(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Remove Method")
}
