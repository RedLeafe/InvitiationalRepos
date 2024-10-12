FROM golang:1.23-bookworm AS build

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN mkdir -p /usr/local/bin/app && go build -v -o /usr/local/bin/app ./...

FROM debian:bookworm-slim

COPY --from=build /usr/local/bin/app/tartarus.moon.mine /usr/local/bin/app/tartarus.moon.mine
COPY ./public /usr/local/bin/app/public
WORKDIR /usr/local/bin/app

ENTRYPOINT ["./tartarus.moon.mine"]