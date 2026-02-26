#### Build
In order to test that this all works as expected, we need to build that Docker image and run it:
```shell
go mod tidy

docker build -t docker-image:cache --platform linux/arm64 .
docker run -p 8080:8080 docker-image:cache
```

#### Local Test
Test simply calling _health_ on running docker service:
```shell
curl http://localhost:8080/health
```

#### Bloom Filter API:

Add keys:
```shell
curl -X POST http://localhost:8080/bloom/add \
  -H "Content-Type: application/json" \
  -d '{"keys": ["key1", "key2", "key3"]}'
```

Check if key exists:
```shell
curl http://localhost:8080/bloom/check/key1
```

Get statistics:
```shell
curl http://localhost:8080/bloom/stats
```

Clear filter:
```shell
curl -X POST http://localhost:8080/bloom/clear
```

#### Metrics:

Prometheus metrics endpoint:
```shell
curl http://localhost:8080/metrics
```

#### Management Commands:

```shell
# Show configuration
./cache mgmt show-config

# Show status of detached service
./cache mgmt status

# Start in detached mode
./cache server start -D

# Stop detached service
./cache server stop -D
```
