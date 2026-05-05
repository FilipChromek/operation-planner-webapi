param (
  $command
)

if (-not $command) { $command = "start" }

$ProjectRoot = "${PSScriptRoot}/.."

$env:OR_PLANNER_API_ENVIRONMENT = "Development"
$env:OR_PLANNER_API_PORT = "8080"
$env:OR_PLANNER_API_MONGODB_USERNAME = "root"
$env:OR_PLANNER_API_MONGODB_PASSWORD = "neUhaDnes"
$env:OR_PLANNER_API_MONGODB_DATABASE = "orp-or-planner"

function mongo {
  docker compose --file ${ProjectRoot}/deployments/docker-compose/compose.yaml $args
}

switch ($command) {
  "openapi" {
    docker run --rm -ti -v ${ProjectRoot}:/local openapitools/openapi-generator-cli generate -c /local/scripts/generator-cfg.yaml
  }
  "start" {
    try {
      mongo up --detach
      go run ${ProjectRoot}/cmd/or-planner-api-service
    } finally {
      mongo down
    }
  }
  "mongo" {
    mongo up
  }
  "test" {
    go test -v ./...
  }
  "docker" {
    docker build -t xchromek/orp-or-planner-webapi:local-build -f ${ProjectRoot}/build/docker/Dockerfile $ProjectRoot
  }
  default {
    throw "Unknown command: $command. Use one of: start, openapi, mongo, test, docker"
  }
}
