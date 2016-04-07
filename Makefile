# Key is that we tell go to work here...
export GOPATH:=$(shell pwd)

define help

Makefile for grunt
  all	     - make everything

  test	     - run tests
  
  benchmarks - run benchmarks

  run        - run a local grunt with an example config (http://localhost:9901)

  grunt      - build the grunt executable into bin/grunt

  demo		 - build a docker that offers basic example
  
  ants		 - build a docker that is aimed to registration utilizig ANTs

  machinelearn
  
  tools      - run 'go get' to install missing tools

endef
export help

help:
	@echo "$$help"

all: grunt

grunt: bin/grunt

bin/grunt: fmt assets $(shell find src/grunt -type f) deps
	go build grunt/...

fmt: $(shell find src/grunt -type f)
	go fmt grunt/...

deps:
	go get grunt/...

assets: src/grunt/assets.go

src/grunt/assets.go: $(shell find assets -type f) bin/go-bindata
	bin/go-bindata ${debug} -prefix assets -o src/grunt/assets.go assets/...

bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

test: grunt
	go test grunt/...

benchmarks: fmt
	go test -run=XXX -bench . -v grunt/...

run: debug = -debug

run: grunt assets
	bin/grunt gruntfile.yml

clean:
	go clean grunt/...
	rm -rf pkg/*

bin/grunt-docker: fmt assets deps
	GOOS=linux GOARCH=amd64 go build -o bin/grunt-docker grunt/...

demo: bin/grunt-docker
	docker build -t pesscara/grunt -f docker/grunt.Dockerfile .

ants:
	docker build -t pesscara/ants -f docker/ants.Dockerfile .

machinelearn:
	docker build -t pesscara/machinelearn -f docker/python.Dockerfile .

.PHONY: ants grunt 
