.PHONY: run build migrate test clean

run:
	go run cmd/main/main.go

build:
	go build -o server.exe cmd/server/main.go

migrate:
	go run cmd/migrate/main.go

clean:
	rm -f server.exe