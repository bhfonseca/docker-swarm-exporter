# Docker Swarm Exporter

A Prometheus exporter for Docker and Docker Swarm metrics.

## Features

This exporter collects the following metrics:

- **Containers**: Running, stopped, and paused containers
- **Services**: Number of services in the swarm
- **Tasks**: Running and desired tasks per service
- **Nodes**: Total and active nodes in the swarm
- **Stacks**: Number of stacks deployed in the swarm

## Installation

### Building from source

```bash
git clone https://github.com/bhfonseca/docker-swarm-exporter.git
cd docker-swarm-exporter
go build
```

### Docker

#### Build locally

```bash
docker build -t docker-swarm-exporter .
docker run -d --name docker-swarm-exporter -p 9323:9323 -v /var/run/docker.sock:/var/run/docker.sock docker-swarm-exporter
```

#### Use pre-built image from GitHub Container Registry

```bash
docker pull ghcr.io/bhfonseca/docker-swarm-exporter:latest
docker run -d --name docker-swarm-exporter -p 9323:9323 -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/bhfonseca/docker-swarm-exporter:latest
```

You can also use specific version tags instead of `latest`. For example:

```bash
docker pull ghcr.io/bhfonseca/docker-swarm-exporter:v1.0.0
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/) (SemVer). Versions are tagged in the format `vX.Y.Z` where:

- `X` is the major version (incompatible API changes)
- `Y` is the minor version (backwards-compatible functionality)
- `Z` is the patch version (backwards-compatible bug fixes)

When using the Docker image, you can specify the version in several ways:

- `latest`: Always use the latest version (from the main branch)
- `vX.Y.Z`: Use a specific version (e.g., `v1.0.0`)
- `vX.Y`: Use the latest patch version of a specific minor version (e.g., `v1.0`)
- `vX`: Use the latest minor and patch version of a specific major version (e.g., `v1`)

You can check the version of the running exporter by:

1. Using the `--version` flag when starting the exporter
2. Looking at the version information on the web interface (http://localhost:9323/)

## Usage

```bash
./docker-swarm-exporter [flags]
```

### Flags

- `--web.listen-address`: Address to listen on for web interface and telemetry (default: ":9323")
- `--web.telemetry-path`: Path under which to expose metrics (default: "/metrics")
- `--docker.socket`: Docker socket path (default: "unix:///var/run/docker.sock")
- `--scrape.timeout`: Timeout for scraping Docker metrics (default: 10s)
- `--version`: Show version information and exit

## Metrics

The exporter exposes the following metrics:

- `docker_containers_running_total`: The number of containers running
- `docker_containers_stopped_total`: The number of containers stopped
- `docker_containers_paused_total`: The number of containers paused
- `docker_images_total`: The number of images
- `docker_services_total`: The number of services
- `docker_tasks_running_total`: The number of tasks running (labeled by service_name)
- `docker_tasks_desired_total`: The number of tasks desired (labeled by service_name)
- `docker_nodes_total`: The number of nodes
- `docker_nodes_active_total`: The number of active nodes
- `docker_stacks_total`: The number of stacks
- `docker_containers_running_all_nodes_total`: The number of containers running on each node (labeled by node_id and node_hostname)
- `docker_containers_running_total_all_nodes`: The total number of containers running across all nodes combined

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'docker-swarm'
    static_configs:
      - targets: ['localhost:9323']
```

## License

MIT
