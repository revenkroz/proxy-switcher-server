# Proxy switcher reverse proxy server

## Description

This is a reverse proxy server that allows you to forward requests to the target server through list of proxies.
Proxy server will switch to the next proxy server from the list if the target server returns a specific http status code.

## Usage

__Note:__ For all examples below we assume that there are two proxy servers and the target server is https://example.com.

### Executable usage

Add environment variables:
```bash
TARGET_URL=https://example.com
PROXY_LIST=http://proxy_1:8081,http://proxy_2:8082
```

And run.
```bash
./proxy-switcher
```

Or use flags:
```bash
./proxy-switcher \
    -target=https://example.com \
    -proxy=http://proxy_1:8081 \
    -proxy=http://proxy_2:8082
```

### List of all flags

* `-target` - target server url
* `-proxy` - proxy server url (can be used multiple times)
* `-trigger-code` - http status code that will trigger proxy switch (default is 429)
* `-listen` - address to listen (default is 0.0.0.0:8888)

### Docker compose usage

```yaml
version: '3.9'

services:
  example_com_proxy:
    image: ghcr.io/revenkroz/proxy-switcher-server:main
    container_name: proxy
    environment:
      TARGET_URL: https://example.com
      PROXY_LIST: http://proxy_1:8081,http://proxy_2:8082
    ports:
        - "8888:8888"
```
