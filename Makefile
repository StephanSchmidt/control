.PHONY: all build install test test-coverage lint check sec upgrade-deps example examples tutorial clean

all: lint sec test build

build: go-imports
	go build -o bin/control .

install: lint sec test build
	go install .

test:
	gotestsum ./...

test-coverage:
	go test -cover

#open:
#	xdg-open index.html 2>/dev/null || open index.html 2>/dev/null || start index.html

examples: build
	@for f in examples/*.txt; do \
		name=$$(basename "$$f" .txt); \
		echo "Building $$name.txt..."; \
		./bin/control --stretch 0.8 --diagram "$$f" --out "examples/$$name.svg" --debug "examples/$$name-debug.json" --font fonts/BerkeleyMono-Condensed.woff2; \
		w=$$(grep -o 'width="[0-9]*"' "examples/$$name.svg" | head -1 | grep -o '[0-9]*'); \
		h=$$(grep -o 'height="[0-9]*"' "examples/$$name.svg" | head -1 | grep -o '[0-9]*'); \
		google-chrome --headless --disable-gpu --hide-scrollbars --screenshot="examples/$$name.png" --window-size=$$w,$$h "examples/$$name.svg"; \
	done

tutorial: build
	@for f in tutorial/*.txt; do \
		name=$$(basename "$$f" .txt); \
		echo "Building tutorial/$$name..."; \
		./bin/control --diagram "$$f" --out "tutorial/$$name.svg"; \
		w=$$(grep -o 'width="[0-9]*"' "tutorial/$$name.svg" | head -1 | grep -o '[0-9]*'); \
		h=$$(grep -o 'height="[0-9]*"' "tutorial/$$name.svg" | head -1 | grep -o '[0-9]*'); \
		google-chrome --headless --disable-gpu --hide-scrollbars --screenshot="tutorial/$$name.png" --window-size=$$w,$$h "tutorial/$$name.svg"; \
	done

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
