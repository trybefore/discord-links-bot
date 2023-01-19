#build binary

FROM golang:1.18-alpine AS builder

RUN apk --update upgrade && apk add --no-cache git make build-base && rm -rf /var/cache/apk/*

ENV GO111MODULE=on

RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG buildOptions

RUN env ${buildOptions} go build -ldflags="-w -s" -o /go/bin/linksbot .

#optimized build

FROM alpine
RUN apk add --no-cache tzdata
RUN apk add --no-cache curl
ENV TZ=Europe/Stockholm
COPY --from=builder /go/bin/linksbot /go/bin/linksbot

RUN apk --update upgrade && apk add --no-cache ca-certificates && update-ca-certificates 2>/dev/null || true && rm -rf /var/cache/apk/*

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=30s --start-period=15s --retries=3 CMD curl --fail http://localhost:8080/health_check || exit 1

ENTRYPOINT ["/go/bin/linksbot"]
