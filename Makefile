.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o bin/yogi ./cmd/main.go

clean:
	rm -rf bin
