# Faro

**Vector search with a good nose.**

Faro is a small vector database written in Go, built from scratch to explore how similarity search works under the hood. It currently supports sequential and parallel exact (brute-force) search in memory. The HTTP API for vectors, on-disk persistence, and HNSW indexing are planned.

> **Status:** early development. Not production-ready. APIs and file formats will change.

## Why this exists

Most people use vector databases as black boxes. Faro exists to understand and document the trade-offs behind them: recall vs. latency, memory cost, index build time, and durability.

## Goals

**Core**
- In-memory collection of vectors with exact (brute-force) k-nearest-neighbor search
- Parallel exact search with cancellation and a benchmark against sequential search
- Distance metrics: cosine, Euclidean (L2), dot product
- Simple HTTP API to insert and query vectors
- On-disk persistence

**Stretch**
- HNSW index for approximate search
- Benchmarks comparing exact vs. approximate search (recall and latency)
- Documentation site about the project and its development

## Exact search

`Collection.Search(query, k)` scans the collection sequentially.
`Collection.SearchParallel(ctx, query, k, workers)` splits the collection into blocks and scans them using up to `workers` goroutines. Each worker sends its best `k` candidates through a channel, and the caller ranks those candidates for the final result. Both methods order neighbors by ascending L2 distance and break ties by ID. Use a positive worker count.

The collection holds a read lock until all workers submit their candidates. Workers check the context between distance calculations. A canceled query waits for its workers to finish and returns the context error. Sorts already in progress run to completion.

## Benchmarks

```sh
go test -race ./...
go test ./core -run '^$' -bench '^BenchmarkSearch$' -benchtime=300ms -count=3 -cpu=4
```

The benchmark uses `k=10` and the same seeded data and query for every implementation within a scenario. The `-cpu=4` flag sets `GOMAXPROCS` to 4. The timed loop includes each complete query, including ranking and merging candidates. Data generation and insertion happen before timing.

Results below are medians of three runs on Linux amd64 with Go 1.27.1 and an AMD Ryzen 7 7735HS. Times are in microseconds per query. Speedup is sequential time divided by the time with four workers.

| Vectors | Dimensions | Sequential | 1 worker | 2 workers | 4 workers | Speedup |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 100 | 16 | 7.8 | 14.4 | 15.0 | 19.0 | 0.41× |
| 100 | 128 | 14.8 | 21.7 | 22.3 | 24.4 | 0.61× |
| 1,000 | 16 | 108.2 | 123.0 | 94.5 | 78.4 | 1.38× |
| 1,000 | 128 | 184.6 | 205.2 | 153.7 | 111.4 | 1.66× |
| 10,000 | 16 | 1,575.8 | 1,620.9 | 1,072.9 | 615.4 | 2.56× |
| 10,000 | 128 | 2,331.8 | 2,294.5 | 1,734.2 | 857.3 | 2.72× |

Sequential search was faster with 100 vectors. With 10,000 vectors and 128 dimensions, four workers cut query time from 2.33 ms to 0.86 ms and increased allocated bytes from 245,856 to 265,024 per query. Across these cases, sequential search made 4 allocations per query and the parallel variants made 12, 19, and 32 with one, two, and four workers respectively.

## Non-goals

Faro is a learning and engineering project, not a production database. It deliberately does not aim to provide sharding, replication, authentication, or a distributed architecture.

## License

Released under the [MIT License](LICENSE).
