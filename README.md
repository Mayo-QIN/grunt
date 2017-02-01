# Grunt

Grunt is a Go server that exposes a REST interface to command line programs.  Grunt is configured through a simple YML file or CLI XML.


## Build

`make grunt`

`make demo # contains the basic image`

## Run

`grunt -p 9901 gruntfile.yml`

Run grunt on port `9901` (the default listening port).

## run demo

Demo
** Need to execute as root or sudo **

    make demo

Was my docker image created?
    docker image
if you can see pesscara/grunt yes
Then to lunch the docker type:

    docker run -d -p 9901:9901 pesscara/grunt

or for ants:

    docker run -i -p 9901:9903 pesscara/ants

You can check that docker is runing utiliing the floowing command

    docker ps 


## REST Endpoints

| endpoint                        | method | parameters       | description                                                 |
|---------------------------------|--------|------------------|-------------------------------------------------------------|
| `/rest/service`                 | GET    | --               | List the services available                                 |
| `/rest/service/{id}             | GET    | `id`             | Detail for service `id`                                     |
| `/rest/service/{id}             | POST   | `id`             | Start a new Job using service `id`                          |
| `/rest/job/{id}`                | GET    | `id`             | Details about a Job                                         |
| `/rest/job/wait/{id}`           | GET    | `id`             | Does not return until the Job completes                     |
| `/rest/job/{id}/file/{filename}` | GET    | `id`, `filename` | Retrieve the file `filename` from the Job specified by `id` |

## Configuration

An example configuration is found in `gruntfile.yml`. A service consists of the following fields:

```
endPoint      -- REST endpoint, e.g. /rest/service/<endPoint>
commandLine   -- Command line to run
                 Some special command line parameters are
                 @value  -- replace this argument with the parameter from the POST
                 <in     -- look for an uploaded file
                 >out    -- the process will generate this file for later download
description   -- description of the endpoint
defaults      -- a hashmap of default values for "@value" parameters
```

This example configuration file exposes 2 endpoints, test and copy. `test` simply echoes the input and can be called like this:

```
curl -X POST  -v --form Message=hi localhost:9991/rest/service/test
```

copy takes input and output files.  `<in` must be provided

```
curl -X POST  -v --form in=@big_file.txt --form out=small_file.txt localhost:9901/rest/service/copy
```

NB: "--form in=@big_file.txt" indicates that curl should send big_file.txt as the form parameter `in`
and the output filename is set to "small_file.txt"

the following example leverages the slicer's CLI xml configureation

```
curl -X POST  -v --form neighborhood=1,1,1 --form inputVolume=@somefile.nii.gz --form outputVolume=somefile.nii.gz localhost:9901/rest/service/MedianImageFilter

```
to retrieve the output data, first find the UUID in the response, and request the file

```
wget localhost:9901/rest/job/eab4ab07-c8f7-44f7-b7d8-87dbd7226ea4/file/out
```

*NB:* we request the output file using the `out` parameter, not the filename we requested

Here is the copy example using jq(http://stedolan.github.io/jq/) to help a bit

```
id=`curl --silent -X POST --form in=@big_file.txt --form out=small_file.txt localhost:9901/rest/service/copy | jq -r .uuid`
wget --content-disposition localhost:9901/rest/job/$id/file/out
```



## Building

make grunt


## Development

These tools are written in the [Go language](https://golang.org/).  Makefile targets are listed by `make help`.

## Example

This is an example of running the `sleep` job for 120 seconds.

```bash
# Start the job and extract the uuid using jq
id=`curl --silent -X POST --form seconds=120 localhost:9901/rest/service/sleep | jq -r .uuid`

# Status of the job
curl -v localhost:9901/rest/job/$id

# Wait for the job to complete
curl -v localhost:9901/rest/job/wait/$id
```

## Docker notes
### Delete the docker images you do not need. 

Once the docker is up an running 

perform:

`docker ps `

or 

`docker ps -a ` if the container was stoped. 

The output will be something like 

    |=> docker ps 
    CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS               NAMES
    aa4f094a4a55        a4a20e41aa68        "/bin/sh -c 'curl ..."   6 minutes ago       Up 6 minutes        9901/tcp            romantic_meitner


select the conaitner ID and perform

```
docker rmi {CONTAINER ID}
```

### Stop all contaiers

```
docker stop $(docker ps -a -q)
```

if you need to stop a specific container 
```
docker stop {CONTAINER ID}
```

### Remove a specific iamge 

Perform 

`docker images`

select the image container ID 

and execute 

```
docker rm {CONTAINER ID}
```


## Acknowledgement 

Supported by the NCI Grant CA160045
