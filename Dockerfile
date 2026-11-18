FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /aweful-dns .

FROM alpine:3.22
RUN apk add --no-cache dnsmasq
COPY --from=build /aweful-dns /usr/local/bin/aweful-dns
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
