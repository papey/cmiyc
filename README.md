# Cache Me If You Can

<p align="center">
    <img src="LOGO.jpg" width="240" alt="logo"/>
</p>

A simple yet to come HTTP load balancer written in Go, configurable via YAML. ~~Supports multiple routes and backends with pluggable load balancing strategies.~~

---

## Features

- Configurable via YAML.
- ~~Route requests based on URL path prefix.~~
- ~~Single backend strategy (always picks the first backend).~~
- ~~Easy to extend with additional load balancing strategies.~~
- ~~cache get requests~~
- no `httputil.ReverseProxy` here.

---

## Configuration

The load balancer is configured using a YAML file. Example:

```yaml
listen: "localhost:8042"
routes:
  /:
    load_balancer_strategy: "single"
    backends:
      - url: "http://localhost:8080"
  /api:
    load_balancer_strategy: "single"
    backends:
      - url: "http://localhost:8080"
```


## Build


```sh
make
```

## Run

```sh
./cmiyc -h
```

## License

See [LICENSE](LICENSE)
