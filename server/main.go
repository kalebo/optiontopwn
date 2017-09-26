package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const schema = `
CREATE TABLE IF NOT EXISTS score (
	ip TEXT,
	victim TEXT NOT NULL,
	perpetrator TEXT NOT NULL,
	timestamp INTEGER
)
`

func main() {
	http.HandleFunc("/recordPwn", handlePwn)
	http.ListenAndServe(":9999", nil)
}

func handlePwn(rw http.ResponseWriter, r *http.Request) {
	recordBytes, _ := ioutil.ReadAll(r.Body)
	record := json.Unmarshal(recordBytes)

	println(r.Body)

}
