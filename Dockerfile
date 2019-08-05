FROM golang:1.12 as build

WORKDIR /go/src/rmazur.io/healthy
COPY . .

RUN go install -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' -v ./...

FROM alpine:latest as ca-certs
RUN apk add -U --no-cache ca-certificates

FROM scratch
COPY --from=build /go/bin/healthyd /healthyd
COPY --from=ca-certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/healthyd"]
