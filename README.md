# Operation Planner – WebAPI

REST API pre plánovanie operácií, správu operačných sál, pacientov a zdravotníckeho personálu.

## Technológie

- **Go 1.23**
- **Gin** – HTTP framework
- **MongoDB** – databáza
- **OpenAPI 3.0** – špecifikácia API (`api/or-planner.openapi.yaml`)
- **Docker** – multi-stage build (scratch image)

## Lokálny vývoj

### Požiadavky

- Go 1.23+
- Docker + Docker Compose
- [openapi-generator-cli](https://openapi-generator.tech/) (pre generovanie kódu)

### Spustenie databázy

```bash
cd deployments/docker-compose
docker compose up -d
```

Spustí MongoDB na porte `27017` a Mongo Express (webové UI) na `http://localhost:8081`.

Prihlasovacie údaje sú v `.env`:

```
OR_PLANNER_API_MONGODB_USERNAME=root
OR_PLANNER_API_MONGODB_PASSWORD=neUhaDnes
```

### Generovanie API kódu

Pred prvým buildом (alebo po zmene OpenAPI špecifikácie) je potrebné vygenerovať server stub:

```bash
docker run --rm \
  -v "${PWD}:/local" \
  openapitools/openapi-generator-cli generate \
  -c /local/scripts/generator-cfg.yaml
go mod tidy
```

### Spustenie servera

```bash
go run ./cmd/or-planner-api-service
```

Server beží na porte definovanom v `OR_PLANNER_API_PORT` (default `8080`).

### Testy

```bash
go test -v ./...
```

## Konfigurácia (env premenné)

| Premenná | Default | Popis |
|----------|---------|-------|
| `OR_PLANNER_API_PORT` | `8080` | Port servera |
| `OR_PLANNER_API_ENVIRONMENT` | `production` | Prostredie |
| `OR_PLANNER_API_MONGODB_HOST` | `mongo` | Hostname MongoDB |
| `OR_PLANNER_API_MONGODB_PORT` | `27017` | Port MongoDB |
| `OR_PLANNER_API_MONGODB_DATABASE` | `orp-or-planner` | Názov databázy |
| `OR_PLANNER_API_MONGODB_USERNAME` | `root` | Používateľ MongoDB |
| `OR_PLANNER_API_MONGODB_PASSWORD` | *(prázdne)* | Heslo MongoDB |
| `OR_PLANNER_API_MONGODB_TIMEOUT_SECONDS` | `5` | Timeout pripojenia (s) |
| `OR_PLANNER_API_MONGODB_COLLECTION_ROOMS` | `rooms` | Kolekcia sál |
| `OR_PLANNER_API_MONGODB_COLLECTION_PATIENTS` | `patients` | Kolekcia pacientov |
| `OR_PLANNER_API_MONGODB_COLLECTION_STAFF` | `staff` | Kolekcia personálu |

## Docker

### Build

```bash
docker build -f build/docker/Dockerfile -t orp-or-planner-webapi .
```

### Spustenie

```bash
docker run -p 8080:8080 \
  -e OR_PLANNER_API_MONGODB_HOST=host.docker.internal \
  -e OR_PLANNER_API_MONGODB_PASSWORD=neUhaDnes \
  orp-or-planner-webapi
```

## CI/CD

GitHub Actions workflow (`.github/workflows/ci.yml`) pri každom push:

1. Vygeneruje API kód z OpenAPI špecifikácie
2. Spustí testy (`go test ./...`)
3. Zbuilduje a pushne Docker image na DockerHub

Pri tagu `v1.*` vytvorí aj semver-tagované verzie image (`1.1.0`, `1.1`, `1`, `latest`).

## Štruktúra projektu

```
api/                    # OpenAPI špecifikácia
cmd/                    # Vstupný bod aplikácie
internal/               # Obchodná logika a handlery
  or_planner/           # Vygenerovaný stub (nepúpravovať ručne)
scripts/                # Konfigurácia generátora
build/docker/           # Dockerfile
deployments/
  docker-compose/       # Lokálna MongoDB + Mongo Express
```

## Autori

Filip Chromek, Dominik Mojto
