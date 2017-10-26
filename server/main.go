package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	Timestamp int `json:"timestamp"`
	ID        int `json:"id"`
}

type Score struct {
	Username         string `json:"username"`
	VictimCount      int    `json:"victim_count"`
	PerpetratorCount int    `json:"perpetrator_count"`
}

type Graph struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

type GraphNode struct {
	ID string `json:"id"`
}

type GraphLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Value  int    `json:"value"`
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
	http.HandleFunc("/", serveGraph)
	http.HandleFunc("/test", serveGraphTest)
	http.HandleFunc("/submit", handlePwn)
	http.HandleFunc("/raw.json", serveRawScores)
	http.HandleFunc("/graph.json", serveGraphScores)
	http.ListenAndServe(":9999", nil)
}

func handlePwn(rw http.ResponseWriter, r *http.Request) {
	recordBytes, _ := ioutil.ReadAll(r.Body)

	record := common.Record{}
	json.Unmarshal(recordBytes, &record)
	if record.Victim == "" || record.Perpetrator == "" || record.Host == "" {
		log.Print("Invalid json record.")
		http.Error(rw, "Invalid json record.", http.StatusBadRequest)
		return
	}

	if strings.Contains(record.Victim, "\\") {
		record.Victim = strings.Split(record.Victim, "\\")[1] // remove the domain prefix for now
	}

	fmt.Printf("Recived pwn from %s (%s): %s took an option on %s\n", record.Host, r.RemoteAddr, record.Perpetrator, record.Victim)
	stmt, _ := db.Prepare("INSERT INTO scoreboard (ip, victim, perpetrator, host, timestamp) VALUES (?, ?, ?, ?, ?)")

	_, err := stmt.Exec(r.RemoteAddr, record.Victim, record.Perpetrator, record.Host, time.Now().Unix())
	if err != nil {
		log.Fatalf("Unable to insert into database: %s", err)
	}

}

func serveRawScores(rw http.ResponseWriter, r *http.Request) {
	var records []TimestampedRecord

	rows, err := db.Query("select rowid, victim, perpetrator, host, timestamp FROM scoreboard")
	for rows.Next() {
		var record TimestampedRecord
		if err := rows.Scan(&record.ID, &record.Victim, &record.Perpetrator, &record.Host, &record.Timestamp); err != nil {
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

func countFrequency(list []common.Record) []GraphLink {
	freq := make(map[string]*GraphLink)

	for _, item := range list {
		key := item.Perpetrator + "->" + item.Victim
		_, exists := freq[key]

		if exists {
			freq[key].Value++
		} else {
			temp := GraphLink{Source: item.Perpetrator, Target: item.Victim, Type: "pwnd", Value: 1}
			freq[key] = &temp
		}

	}

	var result []GraphLink

	for _, item := range freq {
		result = append(result, *item)
	}

	return result
}

func makeNodes(list []common.Record) []GraphNode {
	nodes := make(map[string]*GraphNode)

	for _, item := range list {
		_, existsPerp := nodes[item.Perpetrator]
		_, existsVict := nodes[item.Victim]

		if !existsPerp {
			nodes[item.Perpetrator] = &GraphNode{ID: item.Perpetrator}
		}

		if !existsVict {
			nodes[item.Victim] = &GraphNode{ID: item.Victim}
		}
	}

	var result []GraphNode

	for _, item := range nodes {
		result = append(result, *item)
	}

	return result
}

func serveGraphScores(rw http.ResponseWriter, r *http.Request) {
	var records []common.Record

	rows, err := db.Query("select victim, perpetrator, host FROM scoreboard")
	for rows.Next() {
		var record common.Record
		if err := rows.Scan(&record.Victim, &record.Perpetrator, &record.Host); err != nil {
			log.Fatalf("Unable extract records from DB: %s", err)
		}

		records = append(records, record)

	}
	if rows.Err() != nil {
		log.Fatalf("Unable extract records from DB: %s", err)
	}
	nodes := makeNodes(records)
	links := countFrequency(records)
	graph := Graph{Nodes: nodes, Links: links}

	err = json.NewEncoder(rw).Encode(graph)
	if err != nil {
		log.Fatalf("Unable to marshal records from DB: %s", err)
	}
}

func serveGraph(rw http.ResponseWriter, r *http.Request) {
	http.ServeFile(rw, r, "./server/main.html")
}

func serveGraphTest(rw http.ResponseWriter, r *http.Request) {
	http.ServeFile(rw, r, "./server/testing.html")
}
