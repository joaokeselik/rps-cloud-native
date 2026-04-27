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

## Kora i Kubernetes lokalt

Manifesten i `k8s/` kan koras lokalt med Docker Desktop Kubernetes, minikube eller kind. For Docker Desktop Kubernetes racker det att bygga imagen lokalt med samma namn som manifestet anvander:

```bash
docker build -t rps-api:latest .
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

Om du anvander minikube kan du bygga imagen direkt i minikubes Docker-miljo:

```bash
minikube docker-env
docker build -t rps-api:latest .
```

Om du anvander kind laddar du in imagen i klustret efter bygg:

```bash
kind load docker-image rps-api:latest
```

## Deploy till Kubernetes med Docker Hub-image

1. Uppdatera image-raden i [k8s/app.yaml](/c:/Users/joaok/Desktop/YH%20Akademin%20-%20Cloud%20Native%20Computing/inl%C3%A4mningsuppgift%202/k8s/app.yaml) till din Docker Hub-image, till exempel `ditt-dockerhub-namn/rps-cloud-native:latest`.
2. Justera losenorden i [k8s/mysql-secret.yaml](/c:/Users/joaok/Desktop/YH%20Akademin%20-%20Cloud%20Native%20Computing/inl%C3%A4mningsuppgift%202/k8s/mysql-secret.yaml).
3. Deploya manifesten:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/mysql-secret.yaml
kubectl apply -f k8s/mysql-pvc.yaml
kubectl apply -f k8s/mysql.yaml
kubectl apply -f k8s/app.yaml
```

4. Exponera tjansten lokalt for test:

```bash
kubectl -n rps port-forward svc/rps-api 8000:8000
```

5. Oppna `http://localhost:8000`.

## Inlamning

For sjalva inlamningen behover du normalt:

- URL till ditt privata repo eller dina privata repos
- rapporten i `docs/rapport.md`, uppdaterad med dina egna namn, bilder och repo-lankar

Den har losningen ligger i ett repo, vilket racker for G-kraven. Om du vill dela upp frontend och API i flera repos senare gar det att bygga vidare fran samma grund.
