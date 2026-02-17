.PHONY: all build install test test-coverage lint check sec upgrade-deps example clean

all: build

build: go-imports
	go build -o bin/control .

install: go-imports
	go install .

test:
	gotestsum ./...

test-coverage:
	go test -cover

#open:
#	xdg-open index.html 2>/dev/null || open index.html 2>/dev/null || start index.html

example: build
	./bin/control --stretch 0.8 --diagram examples/scrum.txt --out examples/scrum.svg --debug examples/scrum-debug.json --font fonts/BerkeleyMono-Condensed.woff2
	google-chrome --headless --disable-gpu --screenshot=examples/scrum.png --window-size=1700,1200 examples/scrum.svg

clean:
	go clean -cache -i

go-imports:
	go tool goimports -w .

lint:
	go vet ./...
	go tool staticcheck ./...
	go tool golangci-lint run ./...
	go tool nilaway ./...

sec:
	go tool gosec ./...
	go tool govulncheck ./...

check: lint sec test

upgrade-deps:
	go get -u ./...
	go mod tidy
	gotestsum ./...
