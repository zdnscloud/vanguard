#use golang to build, then copy to alpine
#build in ubuntu will cause lib64 dependency
FROM golang:alpine AS build

RUN mkdir -p /go/src/vanguard
COPY . /go/src/vanguard

WORKDIR /go/src/vanguard
RUN CGO_ENABLED=0 GOOS=linux go build cmd/vanguard/vanguard.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/vanguard/vanguard /usr/local/bin/

EXPOSE 53/udp
EXPOSE 9000/tcp
ENTRYPOINT ["vanguard"]
