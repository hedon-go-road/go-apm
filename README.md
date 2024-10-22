# go-apm

go-apm is a simple apm system for golang.

## Quick Start

Start the base infrastructures and services:

```bash
make docker-up
```

Send some requests to the services:

```bash
curl http://127.0.0.1:30001/order/add?uid=1&sku_id=3&num=1
```

## Some links

- [Jaeger](http://127.0.0.1:16686)
- [Grafana](http://127.0.0.1:3003)
- [Prometheus](http://127.0.0.1:9090)
- [Kibana](http://127.0.0.1:5601)
