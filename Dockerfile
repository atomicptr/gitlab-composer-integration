FROM alpine:3.11 as base

ENV GCI_CACHE_FILE_PATH /cache/gitlab-composer-integration.cache

RUN apk add --no-cache build-base go git

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags '-w -extldflags "-static"'

CMD ["/app/gitlab-composer-integration"]