package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var authToken string
var verificationHost string
var db *sql.DB

func addTables() {
	sqlStmt := `
	create table nodes (id integer not null primary key, val varchar(64) not null );
	create table roots (id integer not null primary key, val varchar(64) not null );
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func main() {
	authToken = os.Args[2]
	verificationHost = os.Args[3]
	var err error
	db, err = sql.Open("sqlite3", "./local.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	http.HandleFunc("/add", addReq)
	http.HandleFunc("/reset", resetReq)
	http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil)
}

// PUT /add
func addReq(w http.ResponseWriter, r *http.Request) {
	vals := parseRequest(r)
	var err []string
	for _, val := range vals {
		if len(val) != 32 {
			err = append(err, val)
		}
		insertNode(val)
		addNode(val)
		deleteNode(val)
	}
	if len(vals) == 0 || len(err) != 0 {
		http.Error(w, "Invalid hash values", http.StatusBadRequest)
	} else {
		js, err := json.Marshal(vals)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func parseRequest(r *http.Request) []string {
	var vals []string
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&vals)
	if err != nil {
		log.Fatal(err)
	}
	return vals
}

func insertNode(val string) {
	q := `INSERT INTO nodes (val) VALUES ($1);`
	db.Exec(q, val)
}

func deleteNode(val string) {
	q := `DELETE FROM nodes WHERE val=$1;`
	db.Exec(q, val)
}

func addNode(val string) {

}

// POST /reset
func resetReq(w http.ResponseWriter, r *http.Request) {
	q := `DROP TABLE IF EXISTS nodes`
	db.Exec(q)
	q = `DROP TABLE IF EXISTS roots`
	db.Exec(q)
	addTables()
}
