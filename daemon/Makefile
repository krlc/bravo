clean:
	sudo kill -9 $(shell cat pid)
	rm -rf log pid daemon

build:
	go build

run: build
	sudo ./daemon

cycle: clean run
