# Hooks

[![build](https://github.com/dotzero/hooks/actions/workflows/ci-build.yml/badge.svg)](https://github.com/dotzero/hooks/actions/workflows/ci-build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dotzero/hooks)](https://goreportcard.com/report/github.com/dotzero/hooks)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/dotzero/hooks/)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/dotzero/hooks/blob/master/LICENSE)

Hooks is a web service to inspect HTTP requests and debug webhooks using a simple web interface.

Data stored in [boltdb](https://github.com/etcd-io/bbolt) (embedded key/value database) files under `BOLT_PATH`.

*The UI was originally created by Jeff Lindsay and his RequestBin service.*

![](https://raw.githubusercontent.com/dotzero/hooks/master/static/img/screenshot.png)

## Running container in Docker

```bash
docker run -d --rm --name hooks -p "8080:8080" dotzero/hooks
```

### Running container with Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: "3"
services:
  hooks:
    image: ghcr.io/dotzero/hooks:latest
    container_name: hooks
    restart: always
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"
    environment:
      HOOKS_HOST: "0.0.0.0"
      HOOKS_PORT: "8080"
      HOOKS_URL: "https://0.0.0.0:8080"
    volumes:
      - hooks_db:/app/var

volumes:
  hooks_db:
```

Run `docker-compose up -d`, wait for it to initialize completely, and visit `http://localhost:8080`

### Build container

```bash
docker build -t dotzero/hooks .
```

## How to run it locally

```
git clone https://github.com/dotzero/hooks
cd hooks
go run .
```

### Command line options

```bash
Usage:
  hooks [OPTIONS]

Application Options:
      --host=        listening address (default: 0.0.0.0) [$HOOKS_HOST]
      --port=        listening port (default: 8080) [$HOOKS_PORT]
      --url=         url to app (default: http://0.0.0.0:8080) [$HOOKS_URL]
      --bolt-path=   parent directory for the bolt files (default: ./var) [$BOLT_PATH]
      --bolt-ttl=    TTL in hours to keep data (default: 48) [$BOLT_TTL_HOURS]
      --static-path= path to website assets (default: ./static) [$STATIC_PATH]
      --tpl-path=    path to templates files (default: ./templates) [$TPL_PATH]
      --tpl-ext=     templates files extensions (default: .html) [$TPL_EXT]
      --verbose      verbose logging
  -v, --version      show the version number

Help Options:
  -h, --help         Show this help message
```

### Environment variables

* `HOOKS_HOST` (*default:* `0.0.0.0`) - listening address
* `HOOKS_PORT` (*default:* `8080`) - listening port
* `HOOKS_URL` (*default:* `http://0.0.0.0:8080`) - url to web UI
* `BOLT_PATH` (*default:* `./var`) - path to BoltDB database (it represents a consistent snapshot of your data)
* `BOLT_TTL_HOURS` (*default:* `48`) - TTL in hours to keep data persistent
* `STATIC_PATH` (*default:* `./static`) - path to web assets
* `TPL_PATH` (*default:* `./templates`) - path to templates
* `TPL_EXT` (*default:* `.html`) - templates files extensions

## License

http://www.opensource.org/licenses/mit-license.php
