target:
	go build -o ./bin/dedup cmd/dedup/main.go
	go build -o ./bin/mover cmd/mover/main.go

clean:
	rm -rf ./bin