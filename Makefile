.PHONY: default

default: build

build: test 
	go build -o app cmd/main.go

test: 
	go test -cover ./...

run: build
	./app

clean:
	rm app