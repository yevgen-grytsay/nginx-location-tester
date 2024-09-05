# Nginx Location Tester

## Docker Compose

З такою конфігурацією запити до сервіса `nginx` будуть здійснюватися з контейнера `tester`:

```yaml
services:
  nginx:
    environment:
      - APP_NGINX_HOST=nginx
      - APP_FETCH_VIA_PROXY=true
```

З такою конфігурацією запити до сервіса `nginx` будуть здійснюватися з браузера:

```yaml
services:
  nginx:
    environment:
      - APP_NGINX_PORT_ON_HOST=8081
      - APP_NGINX_HOST=localhost
      - APP_FETCH_VIA_PROXY=false
```

## Docker

```sh
docker build -t yevhenhrytsai/nginx-location-tester .
docker run --rm -P docker.io/yevhenhrytsai/nginx-location-tester
```

## Resources

- [nxadm/tail](https://github.com/nxadm/tail) -- використовується для читання логу `Nginx`
- [NGINX | A debugging log](https://nginx.org/en/docs/debugging_log.html)
- [nginx: Linux packages](https://nginx.org/en/linux_packages.html#Ubuntu) -- цю процедуру установки довелося проробити,
  щоб у докер-контейнер встановився `nginx-debug`
