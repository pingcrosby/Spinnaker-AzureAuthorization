FROM golang:alpine AS builder

# Fetch dependencies.
# Using go get.
RUN apk update && apk add --no-cache git

# Create appuser.
ENV USER=1001
ENV UID=1001
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/main/demo/

COPY ./svr.go .
COPY ./go.mod .
COPY ./go.sum .

RUN  go mod download github.com/dgrijalva/jwt-go

# Static build required so that we can safely copy the binary over.
RUN CGO_ENABLED=0 go install -ldflags '-extldflags "-static"'


FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -q -r -0 /zoneinfo.zip .


FROM scratch
COPY --from=builder /go/bin/svr /go/bin/svr
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# the timezone data and tls certs from alpine
#ENV ZONEINFO /zoneinfo.zip
#COPY --from=alpine /zoneinfo.zip /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use an unprivileged user.
USER ${USER}:${USER}

# in real world k8s vol
ENTRYPOINT ["/go/bin/svr"]
