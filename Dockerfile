FROM rust:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add musl-dev libressl-dev pkgconfig elfutils perl make git
RUN cargo build --release --bin discord-links-bot

FROM alpine:latest AS runtime
COPY --from=builder /app/target/release/discord-links-bot /usr/local/bin
ENTRYPOINT ["/usr/local/bin/discord-links-bot"]