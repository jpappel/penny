SRC := $(wildcard $(wildcard */*.go)) $(wildcard *.go) $(wildcard api/templates/*.html)

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
