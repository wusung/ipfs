all: build

build:
	go build
	mv cli qfs

install: build
	cp ipfs /usr/local/bin/ipfs

clean:
	rm -f cli/qfs
