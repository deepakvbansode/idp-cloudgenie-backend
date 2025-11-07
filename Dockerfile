# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23-alpine AS build-stage
RUN apk update -q && apk add -q git build-base autoconf automake libtool make g++

##RUN make && make --silent check && make install

COPY . /app

ARG GOARCH
ARG GOOS

ENV GOARCH=${GOARCH}
ENV GOOS=${GOOS}
ENV CGO_ENABLED=1

WORKDIR /app
RUN go mod download
RUN make build


# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Deploy the application binary into a lean image
FROM golang:1.23-alpine AS build-release-stage

WORKDIR /

COPY --from=build-stage /app/bin/application /bin/application

EXPOSE 8080


ENTRYPOINT ["/bin/application"]
CMD ["/bin/application"]