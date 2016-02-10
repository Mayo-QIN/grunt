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
bin/grunt: fmt
	go get -v grunt/...

major: bin/major
bin/major: fmt
	go get -v major/...

fmt:
	go fmt grunt/...
	go fmt major/...

test: grunt major
	go test major/...

benchmarks: vendor fmt
	go test -run=XXX -bench . -v grunt/...

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
