# tk-tests

A minimal Go HTTP service, tested by Testkube Test Workflows.

## The service

| Method | Path          | Description                              |
|--------|---------------|------------------------------------------|
| GET    | `/healthz`    | Health check: `{"status":"ok",...}`       |
| GET    | `/api/items`  | Lists items created so far               |
| POST   | `/api/items`  | Creates an item from `{"name":"..."}`    |

State is in-memory, so it resets on restart. Stdlib only — no dependencies.

```sh
make run                 # listens on :8080 (override with PORT)
make test                # go test ./...
make e2e                 # curl checks, needs the server running
make docker              # builds tk-tests:local
```

## Layout

```
cmd/server/main.go        entrypoint
internal/api/api.go       handlers + routes
internal/api/api_test.go  unit tests (httptest, no network)
test/e2e/api_test.sh      black-box checks against $BASE_URL
scripts/gen-workflows.py  generates the Test Workflows
testkube/                 generated Test Workflow definitions
```

## Testkube

Two workflows, both cloning this repo via `content.git`:

- **`go-unit-tests`** — runs `go vet` and `go test ./...` in `golang:1.24`, uploads
  `coverage.out` as an artifact.
- **`go-api-e2e`** — starts the server as a workflow **service** (`go run ./cmd/server`,
  gated on a `/healthz` readiness probe), then runs `test/e2e/api_test.sh` from a `curl`
  container against `http://{{ services.api.0.ip }}:8080`.

### 1. Create the token secret

The workflows clone with a GitHub token read from a Kubernetes secret in the agent's
namespace. Create it once:

```sh
make tk-secret GITHUB_TOKEN=ghp_xxx
# equivalent to:
# kubectl -n tk-agent create secret generic tk-tests-git --from-literal=token=ghp_xxx
```

A fine-grained PAT with **Contents: read** on this repo is enough. The secret name/key
(`tk-tests-git` / `token`) is referenced by `GIT_SECRET_NAME` / `GIT_SECRET_KEY` in
`scripts/gen-workflows.py`; the username is fixed to `x-access-token`, which GitHub
ignores for PATs.

### 2. Apply and run

```sh
make tk-apply            # regenerate + create/update both workflows
make tk-run              # ...and run them

kubectl testkube run testworkflow go-unit-tests --target testkube.io/source=cloud -f
kubectl testkube run testworkflow go-api-e2e --target testkube.io/source=cloud -f
```

Since the workflows clone from GitHub, push before running — a run tests `main` as it is
on the remote, not the working tree.

### Runners and the `--target` flag

This environment has two runners: the in-cluster agent (`tk-agent`) and a hosted cloud
runner. Only the in-cluster one can read the token secret, so runs are pinned to it with
`--target testkube.io/source=cloud` (a label only the local agent carries). `make tk-run`
does this for you; override with `TK_TARGET=`.

### Fallback: no repo access from the cluster

If the cluster cannot clone the repo, regenerate the workflows with the sources inlined
via `content.files` — no secret and no runner pinning needed, since everything travels in
the workflow spec:

```sh
make tk-apply MODE=files
```

The trade-off is that the workflows then test whatever was inlined at generation time, so
they must be regenerated and re-applied after every source change.

### Troubleshooting

- **`fatal: could not read Username for 'https://github.com'`** — the token secret is
  missing or empty. Run `make tk-secret GITHUB_TOKEN=...`.
- **`Failed to run execution: the runner could not start the execution`**, aborting in
  under a second with nothing in the runner logs — Testkube resolves image metadata from a
  registry before scheduling, so an image that exists only on the node (e.g. via
  `kind load docker-image`) cannot be used. Use a registry-hosted image.
