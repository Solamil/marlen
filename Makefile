BIN := marlen

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BIN) ./cmd/$(BIN)
