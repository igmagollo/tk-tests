# tk-tests

A minimal Go HTTP service, tested by Testkube Test Workflows.

## The service

| Method | Path          | Description                                  |
|--------|---------------|----------------------------------------------|
| GET    | `/healthz`    | Liveness/readiness: `{"status":"ok",...}`     |
| GET    | `/api/items`  | Lists items created so far                    |
| POST   | `/api/items`  | Creates an item from `{"name":"..."}`         |

State is in-memory, so it resets on restart.

```sh
make run                 # listens on :8080 (override with PORT)
make test                # go test ./...
make e2e                 # curl-based checks, needs the server running
make docker              # builds tk-tests:local
```

## Layout

```
cmd/server/main.go       entrypoint
internal/api/api.go      handlers + routes
internal/api/api_test.go unit tests (httptest, no network)
test/e2e/api_test.sh     black-box checks against $BASE_URL
testkube/                Test Workflow definitions
```

## Testkube

Two workflows, both pulling this repo via `content.git`:

- `go-unit-tests` — runs `go vet` and `go test ./...` in `golang:1.24`, uploads `coverage.out` as an artifact.
- `go-api-e2e` — starts the server as a workflow **service** (`go run ./cmd/server`, gated on a
  `/healthz` readiness probe), then runs `test/e2e/api_test.sh` from a `curl` container against
  `http://{{ services.api.0.ip }}:8080`.

```sh
kubectl testkube create testworkflow -f testkube/go-unit-tests.yaml
kubectl testkube create testworkflow -f testkube/go-api-e2e.yaml

kubectl testkube run testworkflow go-unit-tests -f
kubectl testkube run testworkflow go-api-e2e -f
```

Both workflows fetch from GitHub, so changes must be pushed before a run picks them up.
For a private repo, create a token secret in the agent namespace and uncomment the `tokenFrom`
blocks in both YAML files:

```sh
kubectl -n testkube-agent create secret generic tk-tests-git --from-literal=token=<PAT>
```
