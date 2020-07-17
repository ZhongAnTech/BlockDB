.PHONY: blockdb all
all:blockdb
blockdb:
	go build  -o ./blockdb  ./main.go