# Key is that we tell go to work here...
export GOPATH:=$(shell pwd)

define help

Makefile for grunt

  all          - make everything
  test         - run tests
  benchmarks   - run benchmarks
  run          - run a local grunt with an example config (http://localhost:9901)
  grunt        - build the grunt executable into bin/grunt
  demo         - build a docker that offers basic example
  ants         - build a docker that is aimed to registration with ANTs
  machinelearn - build a docker for machine learning
  tools        - run 'go get' to install missing tools

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

test: debug = -debug
test:
	go test -c grunt/...
	go test -v grunt/...

benchmarks: fmt
	go test -run=XXX -bench . -v grunt/...

run: debug = -debug

run: grunt assets
	bin/grunt docker/gruntfile.yml

clean:
	go clean grunt/...
	rm -rf pkg/*

bin/grunt-docker: fmt assets deps
	GOOS=linux GOARCH=amd64 go build -o bin/grunt-docker grunt/...

demo: bin/grunt-docker
	docker build -t pesscara/grunt -f docker/grunt.Dockerfile .

ants:
	docker build -t pesscara/ants -f docker/ants.Dockerfile .

slicer:
	docker build -t pesscara/slicer -f docker/slicer.Dockerfile .

slicer.run:
	docker run --rm -p 9901:9901 -it pesscara/slicer


machinelearn:
	docker build -t pesscara/machinelearn -f docker/python.Dockerfile .

cluster: 
	# docker run --rm -p 8400:8400 -p 8500:8500 -p 8600:53/udp -h node1 --name consul progrium/consul -server -bootstrap -ui-dir /ui
	docker run --rm -p 9901:9901 -it --link consul:consul -e ADVERTISED_PORT=9901 -e ADVERTISED_HOST=192.168.99.100 pesscara/ants 

cluster2: 
	# docker run --rm -p 8400:8400 -p 8500:8500 -p 8600:53/udp -h node1 --name consul progrium/consul -server -bootstrap -ui-dir /ui
	docker run --rm -p 9902:9901 -it --link consul:consul -e ADVERTISED_PORT=9902 -e ADVERTISED_HOST=192.168.99.100 pesscara/machinelearn 

Consul:
	docker run --rm  -p 8400:8400 -p 8500:8500 -p 8600:53/udp -h node1 --name consul progrium/consul -server -bootstrap -ui-dir /ui

rundockers:
	docker run  -p 9918:9901 -d --link consul:consul -e ADVERTISED_PORT=9918 -e ADVERTISED_HOST=192.168.99.100 pesscara/ants
	docker run  -p 9919:9901 -d --link consul:consul -e ADVERTISED_PORT=9919 -e ADVERTISED_HOST=192.168.99.100 pesscara/machinelearn
	docker run  -p 9928:9901 -d --link consul:consul -e ADVERTISED_PORT=9928 -e ADVERTISED_HOST=192.168.99.100 pesscara/ants
	docker run  -p 9929:9901 -d --link consul:consul -e ADVERTISED_PORT=9929 -e ADVERTISED_HOST=192.168.99.100 pesscara/machinelearn
	# docker run  -p 9940:9901 -d --link consul:consul -e ADVERTISED_PORT=9940 -e ADVERTISED_HOST=192.168.99.100 pesscara/machinelearn

docker-ip?=$(shell /usr/local/bin/docker-machine ip default)
naked: bin/grunt
	ADVERTISED_PORT=9902 ADVERTISED_HOST=localhost CONSUL_PORT_8500_TCP_ADDR=${docker-ip} CONSUL_PORT=8500 bin/grunt gf.yml

.PHONY: ants grunt 
