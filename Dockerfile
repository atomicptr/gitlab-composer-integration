FROM alpine:3.11 as base

RUN apk add --no-cache build-base go git

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags '-w -extldflags "-static"'

FROM scratch

ENV GCI_CACHE_FILE_PATH /cache/gitlab-composer-integration.cache

COPY --from=base /app /app
WORKDIR /cache
WORKDIR /app

CMD ["/app/gitlab-composer-integration"]