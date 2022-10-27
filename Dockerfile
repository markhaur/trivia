FROM golang:1.19.0

RUN apt-get update
RUN apt-get install -y curl telnet vim

WORKDIR /usr/src/app

COPY . .
RUN go mod tidy