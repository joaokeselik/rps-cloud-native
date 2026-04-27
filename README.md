# Rock Paper Scissors API

This project is a small but complete solution for the course requirements:

- a Python web service
- database-driven storage of game rounds
- unit tests in CI
- a Docker image that is built and pushed to a registry
- Kubernetes manifests for manual deployment with `kubectl`

The application provides a simple API and web page where users can play Rock, Paper, Scissors. Every game round is stored in the database, and statistics are available through the API.

## Architecture

- `app/` contains the FastAPI application
- `tests/` contains unit tests and API tests
- `.github/workflows/ci.yml` runs tests, builds the image, and pushes it to Docker Hub
- `k8s/` contains manifests for the namespace, MySQL, and the application
- `docs/report.md` contains the project report template

## API Endpoints

- `GET /` serves the web page
- `POST /api/games` plays one round and stores the result
- `GET /api/games` returns the latest rounds
- `GET /api/stats` returns total statistics
- `GET /healthz` provides a health check for Kubernetes

Example API request:

```bash
curl -X POST http://localhost:8000/api/games \
  -H "Content-Type: application/json" \
  -d '{"player_choice":"rock"}'
```

## Run Locally Without Kubernetes

```bash
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

If `DATABASE_URL` is not set, the application uses local SQLite automatically. In Kubernetes, the application uses MySQL through a secret.

## Run Tests

```bash
pytest
```

## CI With GitHub Actions

The workflow does four things:

1. Compiles the Python code with `python -m compileall app tests`
2. Runs unit tests with `pytest`
3. Builds a Docker image
4. Pushes a new Docker image to Docker Hub on pushes to `main` or `master`

To enable the Docker Hub push, add these GitHub repository secrets:

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

The CI workflow pushes the image to:

```text
DOCKERHUB_USERNAME/rps-cloud-native:latest
DOCKERHUB_USERNAME/rps-cloud-native:<commit-sha>
```

## Run in Kubernetes

The manifests in `k8s/` can be used with Docker Desktop Kubernetes, minikube, kind, or another Kubernetes cluster. The application image is pulled from Docker Hub:

```text
keseljoa/rps-cloud-native:latest
```

Make sure CI has pushed the image to Docker Hub before deploying. Then run:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mysql-secret.yaml
kubectl apply -f k8s/mysql-pvc.yaml
kubectl apply -f k8s/mysql.yaml
kubectl apply -f k8s/app.yaml
kubectl -n rps rollout status deployment/mysql
kubectl -n rps rollout status deployment/rps-api
kubectl -n rps port-forward svc/rps-api 8000:8000
```

Keep the `kubectl port-forward` command running while using the app. Then open `http://localhost:8000`.

If you only want to test a local image during development, temporarily change the image line in `k8s/app.yaml` to `rps-api:latest` and build locally:

```bash
docker build -t rps-api:latest .
```

For kind, load the local image into the cluster after building:

```bash
kind load docker-image rps-api:latest
```

If Docker Desktop shows cluster type `kind` but the `kind` command is not available in PowerShell, you can skip `kind load` and continue with `kubectl apply`. Check that the pods are ready before port-forwarding:

```bash
kubectl -n rps get pods
kubectl -n rps rollout status deployment/mysql
kubectl -n rps rollout status deployment/rps-api
```

## Deployment Commands

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mysql-secret.yaml
kubectl apply -f k8s/mysql-pvc.yaml
kubectl apply -f k8s/mysql.yaml
kubectl apply -f k8s/app.yaml
```

Expose the service locally for testing:

```bash
kubectl -n rps port-forward svc/rps-api 8000:8000
```

Keep the terminal running `kubectl port-forward` open. Then open `http://localhost:8000`.

## Submission

For submission, you normally need:

- the URL to your GitHub repository
- the project report in `docs/report.md`, updated with your own name, screenshots, and repository links

This solution uses one repository, which is enough for the requirements. It can be expanded later if you want to split the frontend and API into separate repositories.
