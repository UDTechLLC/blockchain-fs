# Golang image where workspace (GOPATH) configured at /go.
FROM golang:1.9

# FUSE
RUN apt-get update && apt-get install fuse -y

# Copy the local package files to the container’s workspace.
ADD . /go/src/bitbucket.org/udt/wizefs

# Setting up working directory
WORKDIR /go/src/bitbucket.org/udt/wizefs

# Added vendor services will need to be included here
#RUN go get ./vendor
# Get all dependencies
#RUN go get -v ./...

# Just build CLI App and REST API Service
RUN go build -o ./api/wizefs_cli/wizefs_cli -v ./api/wizefs_cli
RUN go build -o ./api/wizefs_mount/wizefs_mount -v ./api/wizefs_mount
RUN go build -o ./api/rest/rest_service -v ./api/rest

# if dev setting will use pilu/fresh for code reloading via docker-compose volume sharing with local machine
#CMD ["./rest/rest_service"]
ENTRYPOINT ["./api/rest/rest_service"]

# REST API Service listens on port 13000.
EXPOSE 13000