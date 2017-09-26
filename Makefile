
.PHONY: ALL

ALL: bin/client bin/server

bin/client: client/main.go
	go build -o bin/client client/main.go

bin/server: server/main.go
	go build -o bin/server server/main.go
