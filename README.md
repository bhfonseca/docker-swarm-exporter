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
git clone https://github.com/devops/docker-swarm-exporter.git
cd docker-swarm-exporter
go build
```

### Docker

```bash
docker build -t docker-swarm-exporter .
docker run -d --name docker-swarm-exporter -p 9323:9323 -v /var/run/docker.sock:/var/run/docker.sock docker-swarm-exporter
```

## Usage

```bash
./docker-swarm-exporter [flags]
```

### Flags

- `--web.listen-address`: Address to listen on for web interface and telemetry (default: ":9323")
- `--web.telemetry-path`: Path under which to expose metrics (default: "/metrics")
- `--docker.socket`: Docker socket path (default: "unix:///var/run/docker.sock")
- `--scrape.timeout`: Timeout for scraping Docker metrics (default: 10s)

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
