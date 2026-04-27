# Rock Paper Scissors API

Detta projekt ar en enkel men komplett losning for G-kraven i uppgiften:

- webbtjanst i Python
- databasdriven lagring av spelomgangar
- enhetstester i CI
- Docker-image som kan byggas och pushas till registry
- Kubernetes-manifest for manuell deploy med `kubectl`

Applikationen erbjuder ett litet API och en enkel webbsida dar anvandaren kan spela sten, sax, pase. Varje spelomgang sparas i databasen och statistik kan hamtas via API.

## Arkitektur

- `app/` innehaller FastAPI-applikationen
- `tests/` innehaller enhetstester och API-tester
- `.github/workflows/ci.yml` kor tester, bygger image och kan pusha till Docker Hub
- `k8s/` innehaller manifest for namespace, MySQL och applikationen
- `docs/rapport.md` ar en svensk rapportmall till inlamningen

## API-endpoints

- `GET /` enkel webbsida for att spela i webblasaren
- `POST /api/games` spelar en omgang och sparar resultatet
- `GET /api/games` visar de senaste omgangarna
- `GET /api/stats` visar total statistik
- `GET /healthz` enkel health check for Kubernetes

Exempel for att spela via API:

```bash
curl -X POST http://localhost:8000/api/games \
  -H "Content-Type: application/json" \
  -d '{"player_choice":"rock"}'
```

## Kora lokalt

```bash
python -m venv .venv
.venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
```

Om du inte anger `DATABASE_URL` anvands lokal SQLite automatiskt. I Kubernetes anvands MySQL via secret.

## Kora tester

```bash
pytest
```

## CI i GitHub Actions

Workflowen gor fyra saker:

1. Kompilerar Python-koden med `python -m compileall app tests`
2. Kor unittester med `pytest`
3. Bygger en Docker-image
4. Pushar en ny Docker-image till Docker Hub vid `push` till `main` eller `master`

For att Docker Hub-pushen ska aktiveras behover du lagga in foljande i GitHub:

- repository secret `DOCKERHUB_USERNAME`
- repository secret `DOCKERHUB_TOKEN`

CI pushar imagen till:

```text
DOCKERHUB_USERNAME/rps-cloud-native:latest
DOCKERHUB_USERNAME/rps-cloud-native:<commit-sha>
```

## Kora i Kubernetes

Manifesten i `k8s/` kan koras med Docker Desktop Kubernetes, minikube eller kind. Applikationen hamtas fran Docker Hub:

```text
keseljoa/rps-cloud-native:latest
```

Se till att CI har hunnit pusha imagen till Docker Hub innan du deployar. Kor sedan:

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

Oppna sedan `http://localhost:8000`.

Om du bara vill testa en lokal image under utveckling kan du tillfalligt andra image-raden i `k8s/app.yaml` till `rps-api:latest` och bygga lokalt:

```bash
docker build -t rps-api:latest .
```

For kind kan en lokal image laddas in i klustret efter bygg:

```bash
kind load docker-image rps-api:latest
```

Om Docker Desktop visar cluster type `kind` men kommandot `kind` inte finns i PowerShell kan du hoppa over `kind load` och fortsatta med `kubectl apply`. Kontrollera sedan att poddarna ar redo innan port-forward:

```bash
kubectl -n rps get pods
kubectl -n rps rollout status deployment/mysql
kubectl -n rps rollout status deployment/rps-api
```

## Deploy-kommandon

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mysql-secret.yaml
kubectl apply -f k8s/mysql-pvc.yaml
kubectl apply -f k8s/mysql.yaml
kubectl apply -f k8s/app.yaml
```

Exponera tjansten lokalt for test:

```bash
kubectl -n rps port-forward svc/rps-api 8000:8000
```

Oppna `http://localhost:8000`.

## Inlamning

For sjalva inlamningen behover du normalt:

- URL till ditt privata repo eller dina privata repos
- rapporten i `docs/rapport.md`, uppdaterad med dina egna namn, bilder och repo-lankar

Den har losningen ligger i ett repo, vilket racker for G-kraven. Om du vill dela upp frontend och API i flera repos senare gar det att bygga vidare fran samma grund.
