# gitlab-composer-integration

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

## TODOs / Limitations

* Fetching data from Gitlab is quite naive in it's current state,
    this will probably not work very well if your Gitlab instance
    has thousands of composer repositories. Not a huge priority for
    me right now as I only have a few hundred composer projects in my
    Gitlab instance.
* Authorization for the composer repository itself. Everyone can see
    your composer projects (at least the ones which the Gitlab token
    can access), pulling will however only work if the user can pull
    the repositories via git.
   
# License

MIT