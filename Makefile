.PHONY: all build-patch patch clean-cache build-gocdr build-bench


all: build

build: patch build-gocdr build-bench

patch: clean-dir build-patch clean-cache

build-patch:
	cd ./goPatch/src && ./make.bash

clean-dir:
	rm -rf ./goPatch/bin ./goPatch/pkg

clean-cache:
	./goPatch/bin/go clean -cache

build-gocdr: 
	cd goCDR && go build
