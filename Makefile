# Key is that we tell go to work here...
export GOPATH:=$(shell pwd)

define help

Makefile for grunt
  all	     - make everything
  vendor     - find dependancies
  test	     - run tests
  benchmarks - run benchmarks

  grunt      - build the grunt executable into bin/grunt
  docker     - build the grunt docker

  tools      - run 'go get' to install missing tools

endef
export help

help:
	@echo "$$help"

all: grunt major

grunt: bin/grunt
bin/grunt: fmt assets
	go get -v grunt/...

major: bin/major
bin/major: fmt
	go get -v major/...

fmt:
	go fmt grunt/...
	go fmt major/...

assets: $(shell find assets -type f) bin/go-bindata
	bin/go-bindata ${debug} -prefix assets -o src/grunt/assets.go assets/...

bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

test: grunt major
	go test major/...

benchmarks: vendor fmt
	go test -run=XXX -bench . -v grunt/...

run: debug = -debug
run: grunt assets
	bin/grunt gruntfile.yml

clean:
	go clean grunt/...
	go clean major/...
	rm -rf pkg/*

docker:
	docker build -t pesscara/grunt -f grunt.Dockerfile .

slicer:
	docker build -t pesscara/slicer -f slicer.Dockerfile .
ants:
	docker build -t pesscara/ants -f ants.Dockerfile .

.PHONY: ants vendor grunt major
