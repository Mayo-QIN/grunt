# Grunt

Grunt is a Go server that exposes a REST interface to command line programs.  Grunt is configured through a simple YML file.

## Build

In the wild use:

``` bash
go get github.com/Mayo-QIN/grunt
```

In a clone of the repo (kudos to the fine [Hellogopher](https://github.com/cloudflare/hellogopher)):

``` bash
make
```

## Run

`grunt gruntfile.yml`

Run grunt on port `9901` (the default listening port).

## Fancy demo

```bash
# Build the grunt docker
docker build -t grunt .

# Run
docker run -d -p 9901:9901 grunt
```

Check the grunt web interface http://localhost:9901

## REST Endpoints

| endpoint                         | method | parameters       | description                                                 |
|----------------------------------|--------|------------------|-------------------------------------------------------------|
| `/rest/service`                  | GET    | --               | List the services available                                 |
| `/rest/service/{id}`             | GET    | `id`             | Detail for service `id`                                     |
| `/rest/service/{id}`             | POST   | `id`             | Start a new Job using service `id`                          |
| `/rest/job/{id}`                 | GET    | `id`             | Details about a Job                                         |
| `/rest/job/wait/{id}`            | GET    | `id`             | Does not return until the Job completes                     |
| `/rest/job/{id}/file/{filename}` | GET    | `id`, `filename` | Retrieve the file `filename` from the Job specified by `id` |

## Configuration

An example configuration is found in `gruntfile.yml`. A service consists of the following fields:

```
endPoint      -- REST endpoint, e.g. /rest/service/<endPoint>
commandLine   -- Command line to run
                 Some special command line parameters are
                 #value  -- replace this argument with the parameter from the POST
                 <in     -- look for an uploaded file
                 >out    -- the process will generate this file for later download
                 ^in     -- uploaded file must be a zip file, extract in a directory (called in) and pass directory name as an argument
                 ~out    -- specify out on the command line as a directory, zip contents for download
description   -- description of the endpoint
defaults      -- a hashmap of default values for "#value" parameters
```

## Endpoints

`grunt` creates [REST endpoints](https://en.wikipedia.org/wiki/Representational_state_transfer) for command line programs.  Endpoints are listed in the `services` array of the YAML configuration file.  Endpoints require a name (`endPoint`) and `commandLine`.  The format of the `commandLine` is an array of options and parameters.  Special one character prefixes to the parameter list tells `grunt` how to map REST requests to the command line of the Endpoint.

Endpoints are configured by adding to the `services` entry in `gruntfile.yml`.  An example is:

```
services:
  - endPoint: toy
    commandLine: ["toy", "#message"]
    description: print message
    defaults:
      message: "Hi From Grunt"
    parameter_descriptions:
      message: This is the message to display at the output
  - endPoint: ball
    commandLine: ["ball", "<input", ">output"]
    description: transforms input to output
    parameter_descriptions:
      input: an input file
      output: the input transformed by the ball operator
```

### Values

If an Endpoint needs a simple value (string, float, integer, etc), the endpoint specifies that by the character `#`.  For example, the `toy` command needs a string as input, and would normally be called as `./toy foo` (where `foo` could be any string).  The corresponding `commandLine` setting would be:

```
commandLine: ["./toy", "#foo"]
```

A REST command to call `toy` with the value `MySpecialValue` would be:

```
curl -X POST -v --form foo=MySpecialValue localhost:9901/rest/service/toy
```

### Input files

Similar to values, input files are prefixed with a `<`.  This tells `grunt` to expect a file to be uploaded.  The uploaded file is saved as the parameter name with out the `<`.  Suppose our `toy` command takes a file as an argument, the corresponding commandLine is now:

```
commandLine: ["./toy", "<input.txt"]
```

[To upload a file using `curl`](http://stackoverflow.com/a/12667839/334619), the `--form` command expects a parameter name and a filename argument starting with `@`, so to upload the local file `local.txt` to `grunt` use:

```
curl -X POST -v --form input.txt=@local.txt localhost:9901/rest/service/toy
```

`grunt` will save `local.txt` to the file `input.txt` on the server before executing the command line `./toy input.txt`.

### Output files

Output files are denoted with a `>` prefix.  This tells `grunt` to expect the command to save a file and the output file should be made available for later download.  Using the `toy` command, suppose it generates a log file.  Because `toy` may generate log files in different formats (perhaps `.txt`, `.xml` or `.json`), `grunt` generates the name of the output file from the REST parameter.  To make this call, `toy log.yml`, this command line is used.

```
commandLine: ["./toy", "<output"]
```

The `output` parameter for the REST call should be the desired filename, e.g. `log.yml`:

```
curl -X POST -v --form output=log.yml localhost:9901/rest/service/toy
```

Notice that value for `output` is `log.yml`.  This REST call will invoke `./toy log.yml` on the server, and the output file `log.yml` can be downloaded after the command is finished as:

```
wget --content-disposition localhost:9901/rest/job/$id/file/out
```

Where `$id` is the id of the Job in `grunt`.

**NB:** `wget` is used because `--content-disposition` honors the output filename passed along by `grunt`, so the output filename would be `log.yml`.

### Input directory

Sometimes a command line program requires the path to a directory as input.  For example, the directory may be full of images to arrange in a [montage](https://www.imagemagick.org/script/montage.php).  The prefix is `^`, and the name of the directory is the rest of the argument.  So if `toy` is expecting a directory called `images`, the command line would be:

```
commandLine: ["./toy", "^images"]
```

The REST command should upload a zip file containing files to go in the images directory.  If the zip file contains a top level directory, it is unzipped in place and renamed to `images` (in the example).  If the zip file has multiple files at the top level, a new directory is created, and the contents of the zip are moved into that directory.

```
curl -X POST -v --form images=@local_images.zip localhost:9901/rest/service/toy
```

### Output directory

A directory full of output files is very similar to a single output file.  Using the `~` prefix, the name of a directory is passed to the command line program.  If `toy` produces a directory full of log files, the command line would be:

```
commandLine: ["./toy", "~logs"]
```

The REST command could specify the name of this directory using the `logs` parameter:

```
curl -X POST -v --form logs=my_log_files localhost:9901/rest/service/toy
```

and `grunt` would create the directory `my_log_files` then run `./toy my_log_files`.  A zip file containing `my_log_files` could be downloaded after the Job is completed by:

```
wget --content-disposition localhost:9901/rest/job/$id/file/logs
```

**NB:** `wget` is used because `--content-disposition` honors the output filename passed along by `grunt`, so the output filename would be `my_log_files.zip`.


### Shims

A [shim](https://en.wikipedia.org/wiki/Shim_(computing)) is a small script that intercepts REST requests and translates them for another command line program.  This may be useful if a command line program needs inputs that `grunt` does not process.  For instance, suppose a command requires a comma-separated list of filenames.  There is no `grunt` parameter prefix to handle that case.  However, we can write a `shim` program in `bash` that looks for files in a directory, formats a comma-separated string, and passes that to another program.  Here's our endPoint definition:

```
services:
  - endPoint: toy
    commandLine: ["shim", "^files"]
```

The `shim` program finds all the filenames in the `files` directory (recall that `^` indicates a zip file upload), adds commas between them an invokes `toy`:

```
#/bin/sh

# call as ./shim files
# will call toy with a comma-separated list of filenames from the files directory.

# create the list
list=$(ls -1 $1 | paste -s -d , - )

# call toy with the list
./toy "$list"
```

# Examples

## Copy Example

The example file `gruntfile.yml` exposes some endpoints. `test` simply echoes the input and can be called like this:

```
curl -X POST  -v --form Message=hi localhost:9991/rest/service/test
```

copy takes input and output files.  `<in` must be provided

```
curl -X POST  -v --form in=@big_file.txt --form out=small_file.txt localhost:9901/rest/service/copy
```

NB: `--form in=@big_file.txt` indicates that curl should send big_file.txt as the form parameter `in`
and the output filename is set to `small_file.txt`

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

## copy-dir example

```bash
# Have a zip file called `test.zip` in the current directory
# Start the job and extract the uuid using jq
id=`curl --silent -X POST --form in=@test.zip --form out=out.zip localhost:9901/rest/service/copy-dir | jq -r .uuid`

# Status of the job
curl -v localhost:9901/rest/job/$id

# Wait for the job to complete
curl -v localhost:9901/rest/job/wait/$id
```


## Sleep example

This is an example of running the `sleep` job for 120 seconds.

```bash
# Start the job and extract the uuid using jq
id=`curl --silent -X POST --form seconds=120 localhost:9901/rest/service/sleep | jq -r .uuid`

# Status of the job
curl -v localhost:9901/rest/job/$id

# Wait for the job to complete
curl -v localhost:9901/rest/job/wait/$id
```

## Acknowledgement 

Supported by the NCI Grant CA160045.
