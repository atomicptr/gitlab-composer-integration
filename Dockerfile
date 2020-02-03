FROM alpine:3.11 as base

RUN apk add --no-cache build-base go git

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags '-w -extldflags "-static"'

FROM scratch

ENV GCI_GITLAB_URL "https://git.yourdomain.com"
ENV GCI_GITLAB_TOKEN "your token here..."
ENV GCI_CACHE_EXPIRE_DURATION "60m"
ENV GCI_CACHE_FILE_PATH "/cache/gitlab-composer-integration.cache"
ENV GCI_HTTP_TIMEOUT "30s"

# default port
EXPOSE 4000

COPY --from=base /app /app
WORKDIR /cache
WORKDIR /app

CMD ["/app/gitlab-composer-integration"]
