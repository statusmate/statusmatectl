# Название бинарного файла
BINARY_NAME = statusmate

# Путь к исходному файлу Go
SRC = ./main.go

# Команды
all: build

build:
	go build -o $(BINARY_NAME) $(SRC)

clean:
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

.PHONY: all build clean run
