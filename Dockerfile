## We specify the base image we need for our
## go application
FROM golang:alpine as build-env

#uncommented because it gave me issues on Raspberry Pi
#RUN apk update && apk add --no-cache git

ENV VERSION=1.0
ENV LOCAL_PORT=8080

## We create an /app directory within our
## image that will hold our application source
## files
RUN mkdir /app

## We copy everything in the root directory
## into our /app directory
COPY . /app

## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app

## we run go build to compile the binary
## executable of our Go program

RUN go build -o main .

# final stage
FROM scratch

COPY --from=build-env . .


## Our start command which kicks off
## our newly created binary executable
CMD ["/app/main"]
