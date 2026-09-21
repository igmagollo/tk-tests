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
scripts/gen-workflows.py  generates the Test Workflows from these sources
testkube/                 generated Test Workflow definitions
```

## Testkube

Two workflows:

- **`go-unit-tests`** — runs `go vet` and `go test ./...` in `golang:1.24`, uploads
  `coverage.out` as an artifact.
- **`go-api-e2e`** — starts the server as a workflow **service** (`go run ./cmd/server`,
  gated on a `/healthz` readiness probe), then runs `test/e2e/api_test.sh` from a `curl`
  container against `http://{{ services.api.0.ip }}:8080`.

```sh
make tk-apply            # regenerate + create/update both workflows
make tk-run              # ...and run them

kubectl testkube run testworkflow go-unit-tests -f
kubectl testkube run testworkflow go-api-e2e -f
```

### Why the sources are inlined

The workflows carry the Go sources and the e2e script inline via `content.files` instead
of the usual `content.git`, because this cluster cannot clone the repo. `scripts/gen-workflows.py`
builds those YAML files from the files on disk, so the repo stays the single source of
truth — **run `make tk-workflows` (or `make tk-apply`) after changing any source**, or the
workflows will keep testing the previous version.

Two other approaches were tried and did not work here:

- `content.git` — the cluster cannot reach/clone this repo.
- A locally built image side-loaded with `kind load docker-image` — Testkube resolves image
  metadata from a registry before scheduling, so an image that exists only on the node
  aborts the execution with *"the runner could not start the execution"*.

If the cluster later gets access to the repo, switching back to git is a small edit to
`scripts/gen-workflows.py`: replace the `content.files` blocks with

```yaml
  content:
    git:
      uri: https://github.com/igmagollo/tk-tests
      revision: main
      tokenFrom:
        secretKeyRef:
          name: tk-tests-git   # kubectl -n tk-agent create secret generic tk-tests-git --from-literal=token=<PAT>
          key: token
```

### Runners

This environment has two runners: the in-cluster agent (`tk-agent`) and a hosted cloud
runner. Both run these workflows, since everything they need is public images plus inlined
content. To pin a run to the local agent:

```sh
kubectl testkube run testworkflow go-unit-tests --target testkube.io/source=cloud
```
