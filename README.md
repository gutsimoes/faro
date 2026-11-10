# Faro

**Vector search with a good nose.**

Faro is a small vector search engine written in Go, built from scratch to explore how similarity search works under the hood. The current milestone implements exact (brute-force) search in memory. HNSW and an experimental comparison against exact search are the next core goals, with HTTP and snapshot persistence planned as supporting features.

> **Status:** early development. Not production-ready. APIs and file formats will change.

## Run the milestone demo

Requires Go 1.23 or newer. No external dependencies.

```sh
go run .
```

The demo inserts four synthetic 2D vectors and finds the three closest to `[0, 0]`:

```text
Faro 0.0-dev — exact L2 search
Vectors: 4 | Query: [0 0] | k: 3
1. origin distance=0.0000
2. near   distance=1.4142
3. middle distance=5.0000
```

These are numeric examples, not text embeddings or semantic search. The demo data is recreated on each run.

## Current implementation

- Fixed-dimension, in-memory collection of vectors and string IDs.
- Euclidean (L2) distance and exact top-k nearest-neighbor search.
- Deterministic ordering: ascending distance, then ID for ties.
- Validation of dimensions, finite coordinates, positive k and unique nonempty IDs.
- Tests and a reproducible benchmark with synthetic vectors.

The `vector` package is independent of the terminal demo, so future HTTP handlers and benchmarks can call the same core. Insert copies its input; duplicate IDs are rejected. Searches return all available vectors if k exceeds the collection size, and an empty result for an empty collection. Collections are not safe for concurrent use.

Search currently scans every vector and sorts all results: O(N × D + N log N) time and O(N) auxiliary memory, for N vectors of dimension D. This is the correctness baseline for evaluating HNSW.

## Validation and benchmarks

```sh
go test ./...
go vet ./...
go build ./...
go test ./vector -run '^$' -bench BenchmarkSearch -benchmem
```

The benchmark uses a fixed random seed, k=10, 1,000 and 10,000 vectors, and 16 and 128 dimensions. Data generation and insertion are excluded from timing. `ns/op` measures time per query; `B/op` and `allocs/op` measure allocations per query. Results depend on the machine and should be compared under the same conditions. No HNSW or recall measurements exist yet.

For the progress presentation: run the demo, show the tests, and run the benchmark. Explain that exact search provides the ground truth for measuring the future HNSW index's recall@k.

## Why this exists

Most people use vector databases as black boxes. Faro exists to understand and document the trade-offs behind them: recall vs. latency, memory cost, index build time, and durability.

## Goals

**Core**
- [x] In-memory collection with exact k-nearest-neighbor search
- [x] Euclidean (L2) distance
- [ ] Cosine and dot-product metrics
- [ ] HNSW index for approximate search
- [ ] Benchmarks comparing exact vs. approximate search (recall and latency)

**Supporting features**
- [ ] Simple HTTP API to insert and query vectors
- [ ] On-disk snapshot persistence

**Stretch**
- Documentation site about the project and its development

## Non-goals

Faro is a learning and engineering project, not a production database. It deliberately does not aim to provide sharding, replication, authentication, or a distributed architecture.

## License

Released under the [MIT License](LICENSE).
