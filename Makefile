BIN       := worng
MODULE    := github.com/KashifKhn/worng
WASM_OUT  := playground/worng.wasm

.PHONY: build test test-unit test-golden test-fuzz generate fmt lint clean wasm install

build:
	go build -o $(BIN) ./cmd/worng

test: test-unit test-golden

test-unit:
	go test ./... -race

test-golden:
	go test ./... -run TestGolden

test-fuzz:
	go test ./internal/lexer/... -fuzz=FuzzLexer -fuzztime=30s
	go test ./internal/parser/... -fuzz=FuzzParser -fuzztime=30s

generate:
	@echo "No code generation in this project."
	@echo "LSP protocol types (internal/lsp/lsproto) will be generated in Phase 2."

fmt:
	gofmt -w .

lint:
	golangci-lint run

clean:
	rm -f $(BIN) $(WASM_OUT) coverage.out

wasm:
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) ./playground

install:
	go install ./cmd/worng
