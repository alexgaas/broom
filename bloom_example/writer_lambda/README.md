#### Build
In order to test that this all works as expected, we need to build that Docker image and run it:
```shell
docker build -t docker-image:writer-lambda --platform linux/arm64 .
docker run -e DRY_RUN="true" -p 9000:8080 docker-image:writer-lambda
```

#### Local Test
The AWS provided base Docker images already come with something called the Runtime Interface Client which takes care of acting as that proxy for you,
allowing the invocation of the function via an HTTP API call.

In order to get our local Lambda to reply with a response, this is what we need to do:
```shell
curl "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'
```

#### Write to Bloom Cache:

Add single key:
```shell
curl -X POST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -H "Content-Type: application/json" \
  -d '{"key": "example.com"}'
```

Add multiple keys:
```shell
curl -X POST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -H "Content-Type: application/json" \
  -d '{"keys": ["url1.com", "url2.com", "url3.com"]}'
```

#### Environment Variables:

| Variable | Description | Default |
|----------|-------------|---------|
| DRY_RUN | Skip actual cache calls | false |
| CACHE_HOST | Cache service URL | http://cache:8080 |


