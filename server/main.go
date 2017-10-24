package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"../common"
)

const schema = `
CREATE TABLE IF NOT EXISTS scoreboard (
	ip TEXT,
	victim TEXT NOT NULL,
	perpetrator TEXT NOT NULL,
	host TEXT,
	timestamp INTEGER NOT NULL
)
`

type TimestampedRecord struct {
	common.Record
	Timestamp string `json:"timestamp"`
	ID        int    `json:"id"`
}

type Score struct {
	Username         string `json:"username"`
	VictimCount      int    `json:"victim_count"`
	PerpetratorCount int    `json:"perpetrator_count"`
}

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
	http.HandleFunc("/raw", serveRawScores)
	http.HandleFunc("/", handlePwn)
	http.ListenAndServe(":9999", nil)
}

func handlePwn(rw http.ResponseWriter, r *http.Request) {
	recordBytes, _ := ioutil.ReadAll(r.Body)

	record := common.Record{}
	json.Unmarshal(recordBytes, &record)
	if record.Victim == "" || record.Perpetrator == "" || record.Host == "" {
		log.Print("Invalid json record.")
		http.Error(rw, "Invalid json record.", http.StatusBadRequest)
	}

	fmt.Printf("Recived pwn from %s (%s): %s took an option on %s\n", record.Host, r.RemoteAddr, record.Perpetrator, record.Victim)
	stmt, _ := db.Prepare("INSERT INTO scoreboard (ip, victim, perpetrator, host, timestamp) VALUES (?, ?, ?, ?, ?)")

	_, err := stmt.Exec(r.RemoteAddr, record.Victim, record.Perpetrator, record.Host, time.Now().Unix())
	if err != nil {
		log.Fatalf("Unable to insert into database: %s", err)
	}

}

func serveScores(rw http.ResponseWriter, r *http.Request) {
}

func serveRawScores(rw http.ResponseWriter, r *http.Request) {
	var records []TimestampedRecord

	rows, err := db.Query("select rowid, victim, perpetrator, host FROM scoreboard")
	for rows.Next() {
		var record TimestampedRecord
		if err := rows.Scan(&record.ID, &record.Victim, &record.Perpetrator, &record.Host); err != nil {
			log.Fatalf("Unable extract records from DB: %s", err)
		}

		records = append(records, record)

	}
	if rows.Err() != nil {
		log.Fatalf("Unable extract records from DB: %s", err)
	}

	err = json.NewEncoder(rw).Encode(records)
	if err != nil {
		log.Fatalf("Unable to marshal records from DB: %s", err)
	}
}
