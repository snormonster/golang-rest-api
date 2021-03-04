# Restful API for a directory listing

This project was an exercise to develop a simple cross-platform program written in Golang, to be deployed in a production Docker container.

The program exposes a RESTful interface which a client can query with a path and expect a json encoded response of a full directory listing at the specified path.


## Prerequisites
Tested successfully on Windows and Linux running:
* Go version 1.16
* Docker version 20.10.2

No special configurations are required.

For more information see:

https://golang.org/doc/install

https://docs.docker.com/get-docker/

## Installation

The project repo can easily be cloned, for example

```git
git clone https://github.com/snormonster/temp.git  
```

All the required files are in the project folder.

## Compiling the program
To build the app to an executable, run the following command in the root directory

```go
go build -o main .
```

Otherwise, to simply run the program
```go
go run main.go
```


## Building Docker Image
The project includes a Dockerfile to get started. 
The Docker build can be kicked of by running the following example:
```docker
docker build -t restful-api .
```
The program will be placed in a folder called /app/, more on this later.

## Running in Docker container
To run the program run in a Docker container is very simple, after building the Image the following command will start it up. However, consider removing the --rm flag to not let Docker automatically clean up the container after it exits.
```docker
docker run --rm -p 8080:8080 -e "NODEBUG=false" -it restful-api

```
The NODEBUG environment variable can be set to true if less debug messages in the log file and Docker terminal is desired.

## API usage
Once the API is started it will post a brief message in the Docker terminal to show that it was started and on which port it listens (inside the container). 

**Getting a directory listing**

A "GET" request can be made to the following endpoint with query paramater "path" to receive a directory listing.
```bash
http://IP_ADDRESS:PORT/ls?path=/app
```
In the above example /app is the path specified with a query parameter called path. 

Once called this will return a json string with all the files and folders at that path, recursively, in order to get a full directory listing.

The path specified above will serve as a good example for the service running in a Docker container because we mounted it in a folder called /app.

Another thing to mention is that the path specified /app/ etc. is a relative path, however a full path can also be entered. For example on windows the following returned a valid result:

```bash
curl http://localhost:10000/ls?path=C:\Users\pierre\Documents\GitHub\temp
```

Note: back slashes as well as forward slashes can be used in the path only where the program is running on a windows system, all others shall only permit forward slashes. Also note that the relative path is relative to the app/working directory.

In the case where a path is not found, a HTTP code 400 (StatusBadRequest) shall be returned in the header together with a human readable string in the body "Specified path not found".

When something disastrous happened while looking up the path a HTTP code 400 in the header and a human readable error message will be returned in the body, nevertheless in such an instance the server will remain operational.

For any other undefined API calls a HTTP code 404 will be returned in the header with a simple description in the body.

**Simple health check**
```bash
http://IP_ADDRESS:8080/health
```
This will return a json string with a status:OK, timestamp of the request, and another timestamp of the last server activity.

{"status":"OK","timestamp":"2021-03-04T11:27:24.368688535Z","lastactivity":"2021-03-04T11:26:58.304483239Z"}

## Unit tests
The project includes a few simple unit tests which can be run by running the following command from a terminal:

```go
go test .
```

## Logs
The program logs info, debug, warn, and error messages. These are printed to stdout and is visible in the Docker terminal, however these also get written to 