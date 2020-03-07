# gitlab-composer-integration
[![](https://api.travis-ci.com/atomicptr/gitlab-composer-integration.svg?branch=master)](https://travis-ci.com/atomicptr/gitlab-composer-integration)
[![](https://goreportcard.com/badge/github.com/atomicptr/gitlab-composer-integration)](https://goreportcard.com/report/github.com/atomicptr/gitlab-composer-integration)

A composer repository for your Gitlab instance.

## Features

* Easy and fast setup, just add your Gitlab URL and a token!
* Automatically detects all repositories which have a composer.json
* Disk persisted caching for faster startup times

## Setup

There are multiple ways to setup this service:

### Docker

```bash
$ docker run --rm \
    -e GCI_GITLAB_URL=https://git.yourdomain.com \
    -e GCI_GITLAB_TOKEN="your token..." \
    -v /path/to/gci-cachedir:/cache \
    -p 4000:4000 \
    atomicptr/gitlab-composer-integration
```

### Compile it yourself

```bash
# Clone the repository
$ git clone git@github.com:atomicptr/gitlab-composer-integration.git
$ cd gitlab-composer-integration
# Build the service
$ go build
# Execute!
$ ./gitlab-composer-integration --gitlab-url=https://git.yourdomain.com \
    --gitlab-token="your token..."
```

## Configuration

You can provide all options as command line argument or as environment variable.

### Gitlab Url (--gitlab-url / GCI_GITLAB_URL) <string> required

The url of your Gitlab instance (for instance https://gitlab.mydomain.com)

### Gitlab Token (--gitlab-token / GCI_GITLAB_TOKEN) <string> required

A Gitlab user token with access to the repositories you want to serve via this service.

1. Go to **https://gitlab.mydomain.com/profile/personal_access_tokens** (User Icon > Settings > Access Token)
2. Enter a name, and the **api** and **read_repository** scope.
3. Create and copy your token!

### Cache Expire Duration (--cache-expire-duration / GCI_CACHE_EXPIRE_DURATION) duration default: 60m

Time until the cache will be invalidated.

### Cache File Path (--cache-file-path / $GCI_CACHE_FILE_PATH) string

Location where the cache file will be stored

### Vendor Whitelist (--vendor-whitelist / $GCI_VENDOR_WHITELIST) []string

A comma seperated list of allowed vendors, for example:

```
$ ./gitlab-composer-integration ... --vendor-whitelist=psr,typo3,myvendor
```

### Port (--port / $GCI_PORT) int default: 4000

Well... the port this service will be running as.

### HTTP Timeout (--http-timeout / $GCI_HTTP_TIMEOUT) duration default: 30s

Timeout for requests to Gitlab.

### No Cache (--no-cache / $GCI_NO_CACHE) boolean default: false

Start without cache.

### HTTP Credentials (--http-credentials / $GCI_HTTP_CREDENTIALS) string

Secure your composer repository from prying eyes by protecting it with a basic HTTP auth. For an example scroll down a bit.

## FAQ

### How can I add a custom repository to composer?

Just add this to your composer.json

```json
{
  "repositories": [
    {
      "type": "composer",
      "url": "https://composer.yourdomain.com"
    }
  ]
}
```

### How can I hide certain projects from the repository?

You can control the available projects via the Gitlab user token provided to the
service. For instance you could create a seperate user for this service (recommended anyway) and
allow/deny access to repositories. 

### How can I add authentication to my repository?

Just use the HTTP Credentials option:

```bash
$ ./gitlab-composer-integration ... --http-credentials="username:password"
```

And within your composer.json you add the credentials like this:

```json
{
  "repositories": [
    {
      "type": "composer",
      "url": "https://username:password@composer.yourdomain.com"
    }
  ]
}
```

[You can read more about HTTP basic authentication with composer here.](https://getcomposer.org/doc/articles/http-basic-authentication.md)

## TODOs / Limitations

* Fetching data from Gitlab is quite naive in it's current state,
    this will probably not work very well if your Gitlab instance
    has thousands of composer repositories. Not a huge priority for
    me right now as I only have a few hundred composer projects in my
    Gitlab instance.
* Support for Gitlab webhooks to invalidate the cache.
   
# License

MIT