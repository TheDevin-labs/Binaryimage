BINARY_COMPILE := binary-image-compile
BINARY_RUN     := binary-image-run
RELEASE        := release

.PHONY: all clean compile-tool run-tool

all: compile-tool run-tool

compile-tool:
	go build -o $(RELEASE)/$(BINARY_COMPILE) ./cmd/compile

run-tool:
	go build -o $(RELEASE)/$(BINARY_RUN) ./cmd/run

clean:
	rm -rf $(RELEASE)

test:
	go test ./...
