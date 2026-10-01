# Faro

**Vector search with a good nose.**

Faro is a small vector database written in Go, built from scratch to explore how similarity search works under the hood. It starts with exact (brute-force) search, an HTTP API and on-disk persistence, with HNSW indexing and benchmarks as the next goal.

> **Status:** early development. Not production-ready. APIs and file formats will change.

## Why this exists

Most people use vector databases as black boxes. Faro exists to understand and document the trade-offs behind them: recall vs. latency, memory cost, index build time, and durability.

## Goals

**Core**
- In-memory collection of vectors with exact (brute-force) k-nearest-neighbor search
- Distance metrics: cosine, Euclidean (L2), dot product
- Simple HTTP API to insert and query vectors
- On-disk persistence

**Stretch**
- HNSW index for approximate search
- Benchmarks comparing exact vs. approximate search (recall and latency)
- Documentation site about the project and its development

## Non-goals

Faro is a learning and engineering project, not a production database. It deliberately does not aim to provide sharding, replication, authentication, or a distributed architecture.

## License

Released under the [MIT License](LICENSE).
