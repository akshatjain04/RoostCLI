.PHONY: all clean

all:
	go mod tidy
	go build -o roostcli main.go 

clean:
	-rm -rf roostcli

linux:
	go mod tidy
	GOOS=linux CGO_ENABLED=0 go build -o roostcli main.go  

windows:
	go mod tidy
	go build .