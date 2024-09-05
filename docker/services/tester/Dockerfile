FROM golang:1.22 AS builder

# ARG TARGETOS
# ARG TARGETARCH

WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 go build -v -o tester

# FROM scratch
FROM alpine
WORKDIR /app
COPY --from=builder /usr/src/app/tester .
COPY public ./public
# COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT [ "/app/tester" ]
