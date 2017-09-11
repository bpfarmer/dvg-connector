package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var authToken string
var verificationHost string
var resetFlag string
var port string
var db *sql.DB

// Node struct
type Node struct {
	Val string
}

func main() {
	log.Println("main():starting up, setting env variables")
	port = fmt.Sprintf(":%s", os.Args[1])
	log.Println("main():port=" + port)
	authToken = os.Args[2]
	log.Println("main():authToken=" + authToken)
	verificationHost = os.Args[3]
	log.Println("main():verificationHost=" + verificationHost)
	resetFlag = os.Args[4]
	log.Println("main():resetFlag=" + resetFlag)
	setupDB()
	defer db.Close()

	log.Println("main():starting server")
	http.HandleFunc("/add", addReq)
	http.HandleFunc("/remove", removeReq)
	//http.HandleFunc("/reset", resetReq)
	http.ListenAndServe(port, nil)
}

func setupDB() {
	log.Println("setupDB():opening db connection")
	var err error
	db, err = sql.Open("sqlite3", "./local.db")
	if err != nil {
		log.Fatal(err)
	}
	if resetFlag == "y" {
		log.Println("setupDB():reset db and migrate")
		q := `DROP TABLE IF EXISTS nodes`
		db.Exec(q)
		q = `DROP TABLE IF EXISTS roots`
		db.Exec(q)
		addTables()
	}
}

func addTables() {
	sqlStmt := `
	create table nodes (id integer not null primary key, val varchar(64) not null );
	create table roots (id integer not null primary key, val varchar(64) not null );
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, sqlStmt)
	}
}

// POST /add
func addReq(w http.ResponseWriter, r *http.Request) {
	log.Println("addReq():received a request to add hashes")
	vals := parseRequest(r)
	var badHashes []string
	var nodes []Node
	log.Println("addReq():inserting parsed values into database")
	for _, val := range vals {
		if len(val) != 64 {
			badHashes = append(badHashes, val)
		} else {
			insertNode(val)
			nodes = append(nodes, Node{Val: val})
		}
	}
	addNodes(nodes)
	if len(vals) == 0 || len(badHashes) != 0 {
		http.Error(w, "Invalid hash values: "+strings.Join(badHashes, ","), http.StatusBadRequest)
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

// POST /remove
func removeReq(w http.ResponseWriter, r *http.Request) {
	log.Println("addReq():received a request to add hashes")
	vals := parseRequest(r)
	var badHashes []string
	var nodes []Node
	log.Println("addReq():inserting parsed values into database")
	for _, val := range vals {
		if len(val) != 64 {
			badHashes = append(badHashes, val)
		} else {
			insertNode(val)
			nodes = append(nodes, Node{Val: val})
		}
	}
	removeNodes(nodes)
	if len(vals) == 0 || len(badHashes) != 0 {
		http.Error(w, "Invalid hash values: "+strings.Join(badHashes, ","), http.StatusBadRequest)
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
	log.Println("parseRequest():parsing request body into json")
	return vals
}

func addNodes(nodes []Node) {
	log.Println("addNodes():making request to verification server")
	url := verificationHost + "/add/"
	log.Println("addNodes():making request to URL=" + url)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodes)

	req, err := http.NewRequest("POST", url, b)
	log.Print("addNodes():setting request body=")
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", authToken)

	log.Println("addNodes():making request to verification instance")
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print("addNodes():receieved response=" + string(body))
}

func removeNodes(nodes []Node) {
	log.Println("addNodes():making request to verification server")
	url := verificationHost + "/remove/"
	log.Println("addNodes():making request to URL=" + url)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(nodes)

	req, err := http.NewRequest("POST", url, b)
	log.Print("addNodes():setting request body=")
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Access-Token", authToken)

	log.Println("addNodes():making request to verification instance")
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(requestDump))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print("addNodes():receieved response=" + string(body))
}

func insertNode(val string) {
	q := `INSERT INTO nodes (val) VALUES ($1);`
	db.Exec(q, val)
}

func deleteNode(val string) {
	q := `DELETE FROM nodes WHERE val=$1;`
	db.Exec(q, val)
}

// POST /reset
func resetReq(w http.ResponseWriter, r *http.Request) {

}
