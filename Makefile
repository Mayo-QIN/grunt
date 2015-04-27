# Key is that we tell go to work here...
export GOPATH:=$(shell pwd)

define help

Makefile for grunt
  all	     - make everything (default)
  deps	     - find dependancies
  test	     - run tests
  benchmarks - run benchmarks

  grunt      - build the grunt executable into bin/grent
  docker     - build the grunt docker

endef
export help

help:
	@echo "$$help"

# Get the dependancies, including those for testing (-t)
deps:
	go get -t -d -v grunt/...


grunt: bin/grunt
bin/grunt: deps fmt
	go install grunt/

fmt:
	go fmt grunt/...

test: deps fmt grunt
	go test -v grunt/...

benchmarks: deps fmt
	go test -run=XXX -bench . -v grunt/...

clean:
	go clean grunt/...

docker:
	docker build -t pesscara/grunt -f grunt.Dockerfile .
