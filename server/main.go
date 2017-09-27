package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"../common"
)

const schema = `
CREATE TABLE IF NOT EXISTS score (
	ip TEXT,
	victim TEXT NOT NULL,
	perpetrator TEXT NOT NULL,
	timestamp INTEGER
)
`

var db *sql.DB

func init() {
	// Init the DB connection and if need be create the table
	var err error
	db, err = sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}

	db.Ping()

	db.Exec(schema)

}

func main() {
	http.HandleFunc("/submit", handlePwn)
	http.HandleFunc("/", handlePwn)
	http.ListenAndServe(":9999", nil)
}

func handlePwn(rw http.ResponseWriter, r *http.Request) {
	recordBytes, _ := ioutil.ReadAll(r.Body)

	record := common.Record{}
	json.Unmarshal(recordBytes, &record)

	fmt.Printf("Recived pwn from %s (%s): %s took an option on %s\n", record.Host, r.RemoteAddr, record.Perpetrator, record.Victim)

}
