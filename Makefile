.PHONY: all build-patch patch clean-cache advocate

all: patch advocate

advocate: 
	cd advocate && go build

patch: clean-dir build-patch clean-cache

build-patch:
	cd ./goPatch/src && ./make.bash

clean-dir:
	rm -rf ./goPatch/bin ./goPatch/pkg

clean-cache:
	./goPatch/bin/go clean -cache

adv: advocate
	cd advocate && ./advocate static -path /home/advocate/Advocate/Experiments/Blocking/main.go -main -output -panic

mkbuiltin:
	cd ./goPatch/src/cmd/compile/internal/typecheck && go run mkbuiltin.go