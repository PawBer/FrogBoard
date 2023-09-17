FROM node:20 as npm

WORKDIR /app

COPY ./. ./

RUN npm install
RUN npm run bundle

FROM golang:1.21-bookworm as build

WORKDIR /app 

RUN apt-get update
RUN apt-get install -y libvips-dev

COPY --from=npm /app/. ./
RUN go mod download

RUN go build ./cmd/frogboard

FROM debian:latest

WORKDIR /app

RUN apt-get update
RUN apt-get install -y libvips

COPY --from=build /app/frogboard ./

RUN ls -la ./

CMD ["./frogboard"]