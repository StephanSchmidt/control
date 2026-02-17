.PHONY: all build install test test-coverage

all: build

build: goimports
	go build -o bin/control .

install: goimports
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

goimports:
	goimports -w .

nilcheck:
	nilaway -exclude-pkgs sumitos/internal/stats ./...

lint:
	go vet ./...
	staticcheck ./...
	# golangci-lint run ./...

sec:
	gosec ./...
	govulncheck ./...

.PHONY: loc

upgrade-deps:
	go get -u ./...
	go mod tidy
	gotestsum ./...

loc:
	loc --exclude ".*\.json|.*\.js"^