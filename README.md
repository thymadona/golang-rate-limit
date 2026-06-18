# golang-rate-limit

# Atomic Go Rate Limiter

A dependency-free, thread-safe backend API rate limiter written from first principles in Go.

## Architectural Foundations

* **Algorithm:** Token Bucket using a **Lazy Evaluation Pattern**. Instead of running a costly background CPU ticker to constantly refresh tokens, it computes the token count dynamically based on the exact time delta elapsed between incoming requests.
* **Concurrency Model:** Uses a two-tiered locking mechanism (`sync.RWMutex` globally for client lookups and a local `sync.Mutex` per client bucket) to handle massive parallel request volumes safely without memory race panics.

## Getting Started

### 1. File Structure

Ensure your code is compiled inside a directory matching your Go module specification:

```text
├── main.go        # Complete server & middleware logic
└── main_test.go   # Parallel concurrency validation test

```

### 2. Run the Server

Execute the entrypoint to boot up the HTTP listener on port `:8080`:

```bash
go run main.go

```

### 3. Verify Limits

The system initializes anonymous network clients with a burst limit of **5 requests**, recovering at **1 token per second**. You can test enforcement by firing a quick burst using `curl`:

```bash
# First 5 calls pass instantly
curl http://localhost:8080/api/data

# The 6th concurrent call drops with a rate limit block
# Output: 429 Too Many Requests: Rate limit exceeded.

```

## Running the Tests

To validate that the atomic locks hold up under immense concurrency pressure without experiencing a state-corruption race condition, run the native race-detector test suite:

```bash
go test -race -v

```
