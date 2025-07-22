BINARY_NAME = st4

SRC = ./main.go

all: build

build:
	go build -o $(BINARY_NAME) $(SRC)

clean:
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run
