# Where do I start?

## 1. Dependencies

In order to run the examples (pipeline folder) you need three things.

  - The **grunt** repository 

  - [Go](https://golang.org)

  - **python** with request library installed or just **curl** (these are two options, feel free to use any tool that interacts with REST api)

## 2. Package your algorithm

To deploy your algorithm as a web app through the python interface you have to:
- Create a **docker file** containing your software dependencies.
- Create a **yml** file describing the functionality of your service, as well inputs and outputs.

### 2a. Creating a docker file

Here is an example docker file (more example docker files can be found in the docker folder of the grunt repository - *.dockerfile extension*):

    FROM pesscara/grunt
    # Maybe I need to install Dev tools for dependancies
    RUN yum install -y wget
    RUN yum install -y git
    RUN yum install zlib-devel bzip2-devel openssl-devel ncurses-devel sqlite-devel readline-devel tk-devel -y
    RUN yum install libitpp atlas blas lapack atlas-devel blas-devel lapack-devel libpng-devel -y
    RUN yum groupinstall "Development tools" -y
    WORKDIR /tmp/
    RUN git clone -b release http://cmake.org/cmake.git
    RUN mkdir /tmp/cmake-build
    WORKDIR /tmp/cmake-build
    RUN ../cmake/bootstrap
    RUN make 
    RUN ./bin/cmake -DCMAKE_BUILD_TYPE:STRING=Release .
    RUN make
    RUN make install
    # Install ANTS
    WORKDIR /tmp/
    RUN git clone https://github.com/stnava/ANTs.git
    RUN mkdir -p /tmp/build
    RUN cd ./build/
    RUN cmake ./ANTs -DCMAKE_BUILD_TYPE=Release  -DBUILD_EXAMPLES=OFF -DBUILD_TESTING=OFF
    RUN make 
    RUN echo export PATH=/tmp/bin:\$PATH >> ~/.bashrc
    RUN echo export ANTSPATH=${ANTSPATH:="/tmp/bin"} >> ~/.bashrc
    # copy .yml file as well te script to run. Need to modify so it works.
    COPY docker/ants.gruntfile.yml /grunt.d/gruntfile.yml
    COPY docker/simpleReg /simpleReg
    COPY docker/n4bias.sh /n4bias.sh
    # What do we run on startup?
    CMD ["/bin/grunt", "/gruntfile.yml"]
    # We expose port 9901 by default
    EXPOSE 9901:9901

####Script Explanation: 

    FROM pesscara/grunt

We use the *pesscara/grunt docker template* to start our builds since is already properly configured for **grunt**
Install the basic libraries. 

    RUN yum install -y wget
    RUN yum install -y git
    RUN yum install zlib-devel bzip2-devel openssl-devel ncurses-devel 


`Note` This example is oriented toward a *centos 7* installation environment. 
Please visit the docker site for more information on how to build a docker file. 

After installing all the necessary libraries  switch user and copy all the necessary code (code that you developed/your app) to the **grunt** directory.

The following code demonstrates how three files are copied (ants.gruntfile.yml, simpleReg, n4bias.sh). 

Two of them are command line executables and one of the them (ants.gruntfile.yml) is the configuration file (please see section 2b).

    COPY docker/ants.gruntfile.yml /grunt.d/gruntfile.yml
    COPY docker/simpleReg /simpleReg
    COPY docker/n4bias.sh /n4bias.sh

Subsequently we run grunt utilizing the yml file we just copied. 

    # What do we run on startup?
    CMD ["/bin/grunt", "/gruntfile.yml"]
    # We expose port 9901 by default

Finally expose the port (here it always should be 9901)

    EXPOSE 9901:9901

### 2b. Creating a yml file

    # Working directory
    # This is the directory path used for working files. If left blank,
    # use a system temp directory
    directory: /data/
    services:
      - endPoint: echo
        commandLine: ["echo", "@message"]
        description: print a message
        defaults:
          message: "Hi From Grunt"
      - endPoint: affine
        commandLine: ["/simpleReg", "-f","<fixed","-m", "<moving","-o", ">registered"]
      - endPoint: n4
        commandLine: ["/n4bias.sh", "-f","<fixed","-o", ">registered"]
        # commandLine: ["/simpleReg", "-d", "@dimension","-f","<fixed","-m", "<moving","-e", ">registered","-w",">warped", "-i",">inverse"]


#### Explanation of the yml file 

A service consists of the following fields:

      endPoint      -- REST endpoint, e.g. /rest/service/<endPoint>
      commandLine   -- Command line to run
                       Some special command line parameters are
                       @value  -- replace this argument with the parameter from the POST
                       <in     -- look for an uploaded file
                       >out    -- the process will generate this file for later download
      description   -- description of the endpoint
      defaults      -- a hashmap of default values for "@value" parameters

So you just have to add an end point to the yml file and describe its use parameters. 

default parameters can also be set. 

### 3c. Start everything 

Created everything, now what?

Build the system and deploy! 

**Build**

    make grunt 
    make demo
    make ants
    make machinelearn

`NOTE` you might need to use sudo (depends on the user permissions)

**Deploy**

To run the docker webapps use

    sudo docker run -d -p 9917:9901 pesscara/machinelearn
    sudo docker run -d -p 9916:9901 pesscara/ants


## 3. Interact with your algorithm

You need **curl** or python and **request**


**curl**
   
    curl -v -X POST --form clusternumber=6 --form imageA=@T1c.nii.gz --form imageB=@1.nii.gz --form output=cluster.nii.gz ril-gpu10:9913/rest/service/kmeansseg

Send two registered images and get a 6 cluster image back

**python**

use the _grunt.py to interact with your webapp. (only requirement is requests library)


Example:

    """
    feature selection demo
    """
    from _grunt import *
    # contact the sevice provider
    g = grunt("http://ril-gpu10:9916")
    # Get the endpoint (The endpoint must exist)
    j=g.classeval
    # specify inputs and outputs
    j.datset="diab.csv"
    j.output="classeval.zip"
    # Execute and wait for the output
    job =j()
    job.wait()
    # Download the output
    job.save_output("output", "/Downloads/")

## 4. Monitor 

Visit the address of the server hosting the docker.

## 5. Microservice registration

Grunt is configured to Connect to a [Consul](https://www.consul.io/) server, if the necessary conditions are met.  First grunt must know where the Consul server is, both host and port.  Grunt also needs to know it's "advertised" host and port, especially if running behind a load balancer or in a docker.  These are the relevant command line flags and environment variables:

- Consul Host: `-consul` or `CONSUL_HOST` or `CONSUL_PORT_8500_TCP_ADDR`
- Consul Port: `-consul-port` or `CONSUL_PORT` or `CONSUL_PORT_8500_TCP_PORT`
- Advertised Host: `-advertised` or `ADVERTISED_HOST`
- Advertised Port: `-advertised-port` or `ADVERTISED_PORT`

At startup, if these variables are set, grunt attempts to register itself with Consul.  In addition, grunt registers health checks, notifying Consul when the number of jobs exceeds the `warnLevel` and `criticalLevel` as configured in the `gruntfile.yml`.

### Example

First, let's start Consul running in a docker:

```
docker run --rm  -p 8400:8400 -p 8500:8500 -p 8600:53/udp -h node1 --name consul progrium/consul -server -bootstrap -ui-dir /ui
```

The UI should be available at http://127.0.0.1:8500 or `http://$(docker-machine ip default):8500` on a Mac.  NB: it is rather helpful to put a `docker` entry in `/etc/hosts`.  For Linux, this would be `127.0.0.1 docker` and on a Mac `192.168.99.100 docker`, or `$(docker-machine ip default)`.  For our purposes, we will assume any Dockers are reachable @ `192.168.99.100` via the $DOCKER_IP variable.

Start a grunt running, and have it connect to Consul:

```
docker run --rm -p 9901:9901 -it --link consul:consul -e ADVERTISED_PORT=9901 -e ADVERTISED_HOST=$DOCKER_IP pesscara/grunt
```

Start a second grunt running, connected to Consul.  NB: the second grunt instance is registered on port `9902`:

```
docker run --rm -p 9902:9901 -it --link consul:consul -e ADVERTISED_PORT=9902 -e ADVERTISED_HOST=$DOCKER_IP pesscara/grunt
```

Visiting the [Consul UI](http://192.168.99.100:8500/ui/#/dc1/services/grunt), we can see two grunt instances registered.  And they can be queried using curl.

```
curl $DOCKER_IP:8500/v1/catalog/service/grunt
[
  {
    "Node": "node1",
    "Address": "172.17.0.2",
    "ServiceID": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
    "ServiceName": "grunt",
    "ServiceTags": [
      "echo",
      "sleep",
      "copy"
    ],
    "ServiceAddress": "192.168.99.100",
    "ServicePort": 9901
  },
  {
    "Node": "node1",
    "Address": "172.17.0.2",
    "ServiceID": "grunt-6384695e-c5b9-4a98-ae50-5e51faeff7ec",
    "ServiceName": "grunt",
    "ServiceTags": [
      "echo",
      "sleep",
      "copy"
    ],
    "ServiceAddress": "192.168.99.100",
    "ServicePort": 9902
  }
]
```

Two grunt services are registered, one at `192.168.99.100:9901` and one at `192.168.99.100:9902`.

#### Grunt service health

Grunt provides health status based on the number of running jobs.  The thresholds are set by `warnLevel` and `criticalLevel` in `gruntfile.yml`.  Consul maintains the health of each grunt service.

```
curl -v  docker:8500/v1/health/service/grunt | jq
[
  {
    "Node": ...
    "Service": {
      "ID": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
      "Service": "grunt",
      "Tags": ...
      "Address": "192.168.99.100",
      "Port": 9901
    },
    "Checks": [
      {
        "Node": "node1",
        "CheckID": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
        "Name": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
        "Status": "passing",
        "Notes": "",
        "Output": "0 jobs",
        "ServiceID": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
        "ServiceName": "grunt"
      },
      ...
```

The status for the grunt services @ `192.168.99.100:9901` is `passing` with `0 jobs`.

Launch a few sleep jobs to get us to the warning threshold:

```
# Do this 6 times
curl -X POST --form seconds=5000 192.168.99.100:9901/rest/service/sleep
```

Grunt updates Consul every 30 seconds, changes may take some time to be registered.

```
curl -v  docker:8500/v1/health/service/grunt | jq
...
   "Checks": [
      {
        "Node": "node1",
        "CheckID": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
        "Name": "grunt-1b9c0e07-86c1-4636-98cd-5bfdbb9d6188",
        "Status": "critical",
        "Notes": "",
        "Output": "6 jobs",
        ...
```

Now Consul reports grunt to be in the critical status, with 6 running jobs.
