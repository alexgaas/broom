#### Docker (compose) configuration for bloom cache:

- **cache** - bloom filter cache service with REST API
- **writer_lambda** - AWS lambda to write data to cache
- **Kong** - API Gateway for routing requests
- **Prometheus** - metrics collection
- **Grafana** - dashboards and visualization

#### Quick Start:

Run simulation:

```shell
docker compose up -d
```

Load test data:
```shell
cd test_data
./load_urls.sh
```

#### Services:

| Service | Port | Description |
|---------|------|-------------|
| cache | 8080 | Bloom filter cache API |
| writer_lambda | 9000 | Lambda for writing to cache |
| kong | 8000 | API Gateway |
| prometheus | 9090 | Metrics |
| grafana | 3000 | Dashboards |

#### API Endpoints:

```shell
# Add keys to bloom filter
curl -X POST http://localhost:8080/bloom/add \
  -H "Content-Type: application/json" \
  -d '{"keys": ["key1", "key2"]}'

# Check if key exists
curl http://localhost:8080/bloom/check/key1

# Get statistics
curl http://localhost:8080/bloom/stats

# Health check
curl http://localhost:8080/health
```

#### Via Writer Lambda:

```shell
curl -X POST http://localhost:9000/2015-03-31/functions/function/invocations \
  -H "Content-Type: application/json" \
  -d '{"keys": ["url1.com", "url2.com"]}'
```


