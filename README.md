#### Bloom filter based caching solution for services.

This example includes:

**bloom** package:
- implementation of **[bloom filter](https://github.com/alexgaas/broom/blob/main/bloom/README.md)** - space-efficient probabilistic data structure for set membership testing

**bloom_example** consists of:
- **[cache](https://github.com/alexgaas/broom/blob/main/bloom_example/cache/README.md)** - backend service (built with _gin-gonic_) for bloom filter cache (built also as docker image)
- **[writer_lambda](https://github.com/alexgaas/broom/blob/main/bloom_example/writer_lambda/README.md)** - AWS lambda (built with AWS Go SDK) to write data to **cache** (built as docker image with Kong as API Gateway)
- **[docker](https://github.com/alexgaas/broom/blob/main/bloom_example/README.md)** to run bloom cache service with monitoring (Prometheus + Grafana)
