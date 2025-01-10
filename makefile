SRC := $(wildcard $(wildcard */*.go)) $(wildcard *.go)

.PHONY: all
all: penny

penny: $(SRC)
	go build .

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	go mod tidy

.PHONY: info
info:
	@echo SRC: $(SRC)
