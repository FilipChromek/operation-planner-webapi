### stage 1: generate API skeleton
FROM openapitools/openapi-generator-cli:latest AS api
WORKDIR /local
COPY api api
COPY scripts scripts
COPY .openapi-generator-ignore .openapi-generator-ignore
RUN docker-entrypoint.sh generate -c /local/scripts/generator-cfg.yaml
ENTRYPOINT ["bash"]

### stage 2: compile binary
FROM golang:1.23 AS build
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || true

COPY internal/ internal/
COPY cmd/ cmd/
COPY --from=api /local/ ./

RUN go mod tidy
RUN go test ./... || true

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -ldflags="-w -s" \
    -installsuffix 'static' \
    -o ./or-planner-srv ./cmd/or-planner-api-service

### stage 3: minimal runtime
FROM scratch
LABEL org.opencontainers.image.authors="Filip Chromek, Dominik Mojto"
LABEL org.opencontainers.image.title="Operating Room Planner WebAPI"
LABEL org.opencontainers.image.description="Operation scheduling, patients, staff and operating rooms management"

ENV OR_PLANNER_API_ENVIRONMENT=production
ENV OR_PLANNER_API_PORT=8080
ENV OR_PLANNER_API_MONGODB_HOST=mongo
ENV OR_PLANNER_API_MONGODB_PORT=27017
ENV OR_PLANNER_API_MONGODB_DATABASE=orp-or-planner
ENV OR_PLANNER_API_MONGODB_COLLECTION_ROOMS=rooms
ENV OR_PLANNER_API_MONGODB_COLLECTION_PATIENTS=patients
ENV OR_PLANNER_API_MONGODB_COLLECTION_STAFF=staff
ENV OR_PLANNER_API_MONGODB_USERNAME=root
ENV OR_PLANNER_API_MONGODB_PASSWORD=
ENV OR_PLANNER_API_MONGODB_TIMEOUT_SECONDS=5

COPY --from=build /app/or-planner-srv ./
EXPOSE 8080
ENTRYPOINT ["./or-planner-srv"]
