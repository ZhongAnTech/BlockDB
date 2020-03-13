.PHONY: blockdb all
all:blockdb
blockdb:
	go build  -o ./build/blockdb  ./main.go