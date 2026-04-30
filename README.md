# Rock Paper Scissors API

This project is a small but complete cloud native solution for the course requirements.

- Python FastAPI web service for Rock, Paper, Scissors
- Go CRUD API for player profiles
- MySQL storage for game rounds
- PostgreSQL storage for player profiles
- frontend that consumes both APIs
- unit tests and compilation in CI
- Docker images built and pushed to Docker Hub
- Kubernetes manifests for manual deployment with `kubectl`
- optional Keel deployment automation for Kubernetes image updates
- PostgreSQL backup CronJob that uploads backups to S3

The main application lets users play Rock, Paper, Scissors. Every game round is stored in MySQL, and statistics are returned through the Python API. The same frontend also consumes the Go Players API, where player profiles can be created, listed, selected, and deleted. When a player is selected, new game rounds store that player's id and name with the result.

## Architecture

- `app/` contains the FastAPI application and frontend
- `go-api/` contains the Go Players CRUD API
- `backup/` contains the PostgreSQL-to-S3 backup image
- `tests/` contains Python unit tests and API tests
- `.github/workflows/ci.yml` runs tests, builds images, and pushes them to Docker Hub
- `k8s/` contains manifests for the namespace, databases, APIs, Keel, and backup CronJob
- `docs/report.md` contains the project report template

## Python API

- `GET /` serves the frontend
- `POST /api/games` plays one round and stores the result, optionally with `player_id` and `player_name`
- `GET /api/games` returns the latest rounds, optionally filtered with `?player_id=1` or `?guest_only=true`
- `GET /api/stats` returns total statistics, optionally filtered with `?player_id=1` or `?guest_only=true`
- `GET /healthz` provides a health check

Example:

```bash
curl -X POST http://localhost:8000/api/games \
  -H "Content-Type: application/json" \
  -d '{"player_choice":"rock"}'
```

## Go Players API

The Go service runs on port `8080` and stores data in PostgreSQL.

- `GET /docs` serves browsable API documentation
- `GET /openapi.json` returns an OpenAPI document
- `GET /healthz` provides a health check
- `GET /api/players` lists players
- `POST /api/players` creates a player
- `GET /api/players/{id}` returns one player
- `PUT /api/players/{id}` updates one player
- `DELETE /api/players/{id}` deletes one player

Example:

```bash
curl -X POST http://localhost:8080/api/players \
  -H "Content-Type: application/json" \
  -d '{"name":"Ada","favorite_move":"rock","rating":1200}'
```

## Run Locally Without Kubernetes

```bash
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

If `DATABASE_URL` is not set, the Python application uses local SQLite automatically. In Kubernetes, it uses MySQL through a secret.

## Run Tests

```bash
pytest
cd go-api
go test ./...
```

## CI With GitHub Actions

The workflow does four things:

1. Compiles the Python and Go code
2. Runs Python and Go unit tests
3. Builds Docker images for the Python API, Go API, and PostgreSQL backup job
4. Pushes new Docker images to Docker Hub on pushes to `main` or `master`

To enable Docker Hub pushes, add these GitHub repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

The CI workflow pushes these images:

```text
DOCKERHUB_USERNAME/rps-cloud-native:latest
DOCKERHUB_USERNAME/rps-cloud-native:<commit-sha>
DOCKERHUB_USERNAME/rps-cloud-native-go-api:latest
DOCKERHUB_USERNAME/rps-cloud-native-go-api:<commit-sha>
DOCKERHUB_USERNAME/rps-cloud-native-postgres-backup:latest
DOCKERHUB_USERNAME/rps-cloud-native-postgres-backup:<commit-sha>
```

## Run in Kubernetes

The manifests in `k8s/` can be used with Docker Desktop Kubernetes, minikube, kind, or another Kubernetes cluster. The application images are pulled from Docker Hub:

```text
keseljoa/rps-cloud-native:latest
keseljoa/rps-cloud-native-go-api:latest
```

Make sure CI has pushed the images to Docker Hub before deploying. Then run:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mysql-secret.yaml
kubectl apply -f k8s/mysql-pvc.yaml
kubectl apply -f k8s/mysql.yaml
kubectl apply -f k8s/app.yaml
kubectl apply -f k8s/postgres-secret.yaml
kubectl apply -f k8s/postgres-pvc.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/players-api.yaml
kubectl apply -f k8s/keel.yaml
kubectl -n rps rollout status deployment/mysql
kubectl -n rps rollout status deployment/rps-api
kubectl -n rps rollout status deployment/postgres
kubectl -n rps rollout status deployment/players-api
kubectl -n keel rollout status deployment/keel
```

Expose both services locally:

```bash
kubectl -n rps port-forward svc/rps-api 8000:8000
kubectl -n rps port-forward svc/players-api 8080:8080
```

Keep both `kubectl port-forward` commands running while using the app. Open `http://localhost:8000`. The Go API documentation is available at `http://localhost:8080/docs`.

## Local Image Testing

If you only want to test a local Python image during development, temporarily change the image line in `k8s/app.yaml` to `rps-api:latest` and build locally:

```bash
docker build -t rps-api:latest .
```

For kind, load the local image into the cluster after building:

```bash
kind load docker-image rps-api:latest
```

If Docker Desktop shows cluster type `kind` but the `kind` command is not available in PowerShell, you can skip `kind load` and continue with `kubectl apply`.

## Restart Kubernetes After App Changes

After changing the Python API or frontend template, build and push a fresh Docker image, then restart only the `rps-api` deployment:

```bash
docker build -t keseljoa/rps-cloud-native:latest .
docker push keseljoa/rps-cloud-native:latest
kubectl -n rps rollout restart deployment/rps-api
kubectl -n rps rollout status deployment/rps-api --timeout=240s
kubectl -n rps get pods -o wide
```

Start the port-forwards again in two separate terminals:

```bash
kubectl -n rps port-forward svc/rps-api 8000:8000
```

```bash
kubectl -n rps port-forward svc/players-api 8080:8080
```

Open `http://localhost:8000` and hard refresh the browser with `Ctrl + F5` if the old frontend is still cached.

## Automatic Kubernetes Updates With Keel

Keel is included as an optional Kubernetes automation component. It watches the Docker Hub image tags used by the app deployments and rolls the deployment when the same `latest` tag points to a new image digest.

The app deployments are annotated with:

```yaml
keel.sh/policy: force
keel.sh/trigger: poll
keel.sh/match-tag: "true"
keel.sh/pollSchedule: "@every 2m"
```

Deploy Keel:

```bash
kubectl apply -f k8s/keel.yaml
kubectl -n keel rollout status deployment/keel --timeout=240s
kubectl -n keel get pods
```

After GitHub Actions pushes a new `latest` image to Docker Hub, Keel should notice it within a few minutes and update the matching deployment. You can watch Keel and the app rollout with:

```bash
kubectl -n keel logs deployment/keel --tail=100 -f
kubectl -n rps get pods -w
```

With Keel running, you usually do not need to run `kubectl -n rps rollout restart deployment/rps-api` manually after a successful image push. If you want to disable Keel, remove it with:

```bash
kubectl delete namespace keel
```

## PostgreSQL Backup to S3

The file `k8s/postgres-backup-cronjob.yaml` defines a CronJob that runs every night at 02:00 and uploads a compressed PostgreSQL dump to S3.

Create the S3 secret with your own values:

```bash
kubectl -n rps create secret generic s3-backup-secret \
  --from-literal=AWS_ACCESS_KEY_ID=replace-me \
  --from-literal=AWS_SECRET_ACCESS_KEY=replace-me \
  --from-literal=AWS_DEFAULT_REGION=eu-north-1 \
  --from-literal=S3_BUCKET=replace-me \
  --from-literal=S3_PREFIX=rps-cloud-native/postgres
```

Deploy the CronJob:

```bash
kubectl apply -f k8s/postgres-backup-cronjob.yaml
```

Test the backup manually:

```bash
kubectl -n rps create job --from=cronjob/postgres-s3-backup postgres-s3-backup-manual
kubectl -n rps logs job/postgres-s3-backup-manual
```

