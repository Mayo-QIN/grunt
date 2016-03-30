# Where do I start?

## 1. Dependencies

In order to run the examples (pipeline folder) you need three things.
    - The **grunt** repository 
    - Go (programming language installed)
    - **python** with request library installed or just **curl**

## 2. Package your algorithm

To deploy your algorithm as a web app through the python interface you have to:
- Create a **docker file** containing the dependencies of you software
- Create a **yml** file describing the functionality of your service, as well inputs and outputs.

### 2a. Creating a docker file

Here is an example docker file (more example docker files can be found in the docker folder of the grunt repository - *.dockerfile extension*):

    FROM pesscara/grunt
    USER root
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
    RUN make -j4
    RUN ./bin/cmake -DCMAKE_BUILD_TYPE:STRING=Release .
    RUN make
    RUN make install
    # Install ANTS
    WORKDIR /tmp/
    RUN git clone https://github.com/stnava/ANTs.git
    RUN mkdir -p /tmp/build
    RUN cd ./build/
    RUN cmake ./ANTs -DCMAKE_BUILD_TYPE=Release  -DBUILD_EXAMPLES=OFF -DBUILD_TESTING=OFF
    RUN make -j8
    RUN echo export PATH=/tmp/bin:\$PATH >> ~/.bashrc
    RUN echo export ANTSPATH=${ANTSPATH:="/tmp/bin"} >> ~/.bashrc
    # copy .yml file as well te script to run. Need to modify so it works.
    USER grunt
    COPY docker/ants.gruntfile.yml /grunt/gruntfile.yml
    COPY docker/simpleReg simpleReg
    COPY docker/n4bias.sh n4bias.sh
    # What do we run on startup?
    CMD ["/grunt/grunt", "gruntfile.yml"]
    # We expose port 9901 by default
    EXPOSE 9901:9901

Script analysis: 

    FROM pesscara/grunt
    USER root

We use the pesscara/grunt template to start all our builds since this already contains grunt and proper configuration.

Install the basic libraries. 

    RUN yum install -y wget
    RUN yum install -y git
    RUN yum install zlib-devel bzip2-devel openssl-devel ncurses-devel 


`Note` This example is oriented toward a *centos 7* installation environment. 
Please visit the docker site on more information on how to build a docker file. 

After installing all the necessary libraries  switch user and copy all the necessary code (code that you developed/your app) to the **grunt** directory.

In the following code i copy the following three files: ants.gruntfile.yml, simpleReg,n4bias.sh. Two of them are command line executables and one of the them (ants.gruntfile.yml) is the configuration file (please see section 2b).

    USER grunt
    COPY docker/ants.gruntfile.yml /grunt/gruntfile.yml
    COPY docker/simpleReg simpleReg
    COPY docker/n4bias.sh n4bias.sh

Subsequently we run grunt utilizing the yml file we just copied. 

    # What do we run on startup?
    CMD ["/grunt/grunt", "gruntfile.yml"]
    # We expose port 9901 by default

Finally expose the port (here it always should be 9901)

    EXPOSE 9901:9901

### 2b. Creating a yml file

    # Working directory
    # This is the directory path used for working files. If left blank,
    # use a system temp directory
    directory: /grunt/grunt-tmp/
    services:
      - endPoint: echo
        commandLine: ["echo", "@message"]
        description: print a message
        defaults:
          message: "Hi From Grunt"
      - endPoint: affine
        commandLine: ["./simpleReg", "-f","<fixed","-m", "<moving","-o", ">registered"]
      - endPoint: n4
        commandLine: ["./n4bias.sh", "-f","<fixed","-o", ">registered"]
        # commandLine: ["./simpleReg", "-d", "@dimension","-f","<fixed","-m", "<moving","-e", ">registered","-w",">warped", "-i",">inverse"]


Explanation of the yml file 

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

Created everything now what?

Build the system and deploy! 

**Build**

    make grunt 
    make demo
    make ants
    make machinelearn
    you might need to use sudo (depends on the user permissions)

**Deploy**

To run the docker webapps use

    sudo docker run -d -p 9917:9901 pesscara/machinelearn
    sudo docker run -d -p 9916:9901 pesscara/ants


Need **go** and **docker**

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
    j.datset="/Users/m112447/Documents/TestData/diab.csv"
    j.output="classeval.zip"
    # Execute and wait for the output
    job =j()
    job.wait()
    # Download the output
    job.save_output("output", "/Users/m112447/Downloads/")

## 4. Monitor 

Visit the address of the server hosting the docker. 





