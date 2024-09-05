FROM golang:1.22 AS builder
WORKDIR /usr/src/app
COPY . .
RUN CGO_ENABLED=0 go build -v -o tester


FROM ubuntu:latest
RUN apt-get update && apt-get install -y supervisor
RUN apt-get update && apt-get install -y \
    curl \
    gnupg2 \
    ca-certificates \
    lsb-release \
    ubuntu-keyring
RUN curl https://nginx.org/keys/nginx_signing.key | gpg --dearmor \
    | tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null \
    && gpg --dry-run --quiet --no-keyring --import --import-options import-show /usr/share/keyrings/nginx-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] \
    http://nginx.org/packages/ubuntu `lsb_release -cs` nginx" \
    | tee /etc/apt/sources.list.d/nginx.list \
    && apt-get update && apt-get install -y nginx

# Tester
COPY --from=builder /usr/src/app/tester /app/
COPY ./public /app/public

# Nginx
COPY ./docker/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY ./docker/nginx/share/web_files /usr/share/nginx/web_files

# Supervisord
COPY ./docker/supervisord/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY ./docker/supervisord/nginx.conf /etc/supervisor/conf.d/nginx.conf
COPY ./docker/supervisord/tester.conf /etc/supervisor/conf.d/tester.conf

EXPOSE 8080
EXPOSE 80

CMD ["/usr/bin/supervisord"]
