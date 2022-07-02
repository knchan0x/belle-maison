###########################################################
##### stage 1: build image and name stage as builder ######
###########################################################

# use golang-alpine as base image
FROM golang:1.18.0-alpine3.15 AS builder

# fix missing Git command
RUN apk add git

ENV GO115MODULE=on \ 
    CGO_ENABLED=0 \ 
    GOOS=linux \ 
    GOARCH=amd64

# copy the build content specified to dest's /go/Github/belle-masion (golang-alpine's root folder is /go)
COPY .  Github/belle-masion

# move to working directory
WORKDIR Github/belle-masion/cmd/scheduler

# build executable
RUN go build -o scheduler .

###########################################################
# stage 2: copy the executable and build the actual image #
###########################################################

# use alpine as base image
FROM alpine:3.15

# copy from builder stage's working directory to dest's root (alpine's root folder is /)

# copy executable and config file to project root
COPY --from=builder \ 
        /go/Github/belle-masion/cmd/scheduler/scheduler \ 
        /go/Github/belle-masion/config.yaml \ 
        ./

# run executable
CMD ["./scheduler"]