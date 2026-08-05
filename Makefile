.PHONY: all build-patch patch clean-cache advocate

all: build

advocate: 
	cd advocate && go build

build: patch advocate

patch: clean-dir build-patch clean-cache

build-patch:
	cd ./goPatch/src && ./make.bash

clean-dir:
	rm -rf ./goPatch/bin ./goPatch/pkg

clean-cache:
	./goPatch/bin/go clean -cache

adv: advocate
	cd advocate && ./advocate analysis -path /home/erik/Arbeit/examples/blocking/main.go -main -output -panic

mkbuiltin:
	cd ./goPatch/src/cmd/compile/internal/typecheck && go run mkbuiltin.go
