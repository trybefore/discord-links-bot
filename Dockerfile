FROM golang:1.18-alpine

WORKDIR /app 

COPY go.mod ./ 
COPY go.sum ./ 
RUN go mod download 

COPY *.go ./ 

RUN go build -o /fxdiscordbot

EXPOSE 8080 

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "curl", "--fail", "http://localhost:8080", "||", "exit 1"]

CMD ["/fxdiscordbot"]