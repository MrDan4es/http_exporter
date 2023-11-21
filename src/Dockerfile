FROM golang:1.21.4-alpine3.18 AS build

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -o /godocker


FROM scratch

WORKDIR /

COPY --from=build /godocker /godocker
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/godocker"]
