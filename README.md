# Grunt

Grunt is a Go server that exposes a REST interface to command line programs.  Grunt is configured through a simple YML file.

Grunt development was sponsored by NCI Grant CA160045.

## Build

`make grunt`

## Run

`grunt -p 9901 gruntfile.yml`

Run grunt on port `9901` (the default listening port).

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


## ToDo

grunt can not write `.nii.gz` files correctly, comes out at `*.gz` without the `nii` part.

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

##Notes 
###Delete al docker images

    # Delete all containers
    docker rm $(docker ps -a -q)
    # Delete all images
    docker rmi $(docker images -q)

##Acknowledgement 

Supported by the NCI Grant CA160045
