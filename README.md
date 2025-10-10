# Cache Me If You Can

<p align="center">
    <img src="LOGO.jpg" width="240" alt="logo"/>
</p>

A HTTP reverse proxy & load balancer written in Go, configurable via YAML. Supports multiple routes and backends ~~with pluggable load balancing strategies.~~

---

## Features

- Configurable via YAML.
- Route requests based on URL path prefix matching.
- Supports graceful shutdown.
- Single backend strategy (always picks the first backend).
- Configurable per route, load balancing strategies.
- Per configured route cache usage & configuration.
- no `httputil.ReverseProxy` here.

---

## Configuration

The load balancer is configured using a YAML file. Example:

```yaml
listen: "localhost:8042"
routes:
  /:
    load_balancer_strategy: "single"
    cache:
      enabled: false
    backends:
      - url: "http://localhost:8080"
  /api:
    load_balancer_strategy: "single"
    cache:
      enabled: true
      max_size: 500
      max_entry_size: 1
      ttl: 60
    backends:
      - url: "http://localhost:8081"
```


## Build


```sh
make
```

## Run

```sh
./cmiyc -h
```

## Run Test Suite

```sh
make validate
```

## License

See [LICENSE](LICENSE)
