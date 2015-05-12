# Key is that we tell go to work here...
export GOPATH:=$(shell pwd)
export GOROOT:=$(shell go env GOROOT)
export PATH:=$(shell pwd)/bin:${PATH}

define help

Makefile for grunt
  all	     - make everything (default)
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

grunt: bin/grunt
bin/grunt: vendor fmt
	bin/gb build grunt

major: bin/major
bin/major: vendor fmt
	bin/gb build major

fmt:
	go fmt grunt/...
	go fmt major/...

test: vendor fmt grunt
	bin/gb test major/...

benchmarks: vendor fmt
	go test -run=XXX -bench . -v grunt/...

clean:
	go clean grunt/...
	rm -rf pkg/*

docker:
	docker build -t pesscara/grunt -f grunt.Dockerfile .

slicer:
	docker build -t pesscara/slicer -f slicer.Dockerfile .
ants:
	docker build -t pesscara/ants -f ants.Dockerfile .

tools: bin/gb

bin/gb:
	go get -u github.com/constabulary/gb/...


# Get the dependancies, including those for testing (-t)
vendor: tools
	bin/gb vendor \
	github.com/satori/go.uuid \
	gopkg.in/yaml.v2 \
# Get the dependancies, including those for testing (-t)
vendor:
	bin/gb vendor \
	github.com/satori/go.uuid \
	gopkg.in/yaml.v2 \
	github.com/codegangsta/cli \
	gopkg.in/tylerb/graceful.v1 \
	github.com/Sirupsen/logrus \
	gopkg.in/mgo.v2 \
	code.google.com/p/gorest \


.PHONY: ants vendor grunt major
