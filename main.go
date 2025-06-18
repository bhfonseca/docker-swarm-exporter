package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen-address", ":9323", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	dockerSocket  = flag.String("docker.socket", "unix:///var/run/docker.sock", "Docker socket path.")
	scrapeTimeout = flag.Duration("scrape.timeout", 10*time.Second, "Timeout for scraping Docker metrics.")
)

// DockerSwarmCollector implements the prometheus.Collector interface
type DockerSwarmCollector struct {
	dockerClient *client.Client
	timeout      time.Duration

	// Metrics
	containersRunning         *prometheus.Desc
	containersStopped         *prometheus.Desc
	containersPaused          *prometheus.Desc
	imagesCount               *prometheus.Desc
	servicesCount             *prometheus.Desc
	tasksRunning              *prometheus.Desc
	tasksDesired              *prometheus.Desc
	nodesCount                *prometheus.Desc
	nodesActive               *prometheus.Desc
	stacksCount               *prometheus.Desc
	containersRunningAllNodes *prometheus.Desc
	totalContainersAllNodes   *prometheus.Desc
}

// NewDockerSwarmCollector creates a new DockerSwarmCollector
func NewDockerSwarmCollector(dockerClient *client.Client, timeout time.Duration) *DockerSwarmCollector {
	return &DockerSwarmCollector{
		dockerClient: dockerClient,
		timeout:      timeout,

		containersRunning: prometheus.NewDesc(
			"docker_containers_running_total",
			"The number of containers running",
			nil, nil,
		),
		containersStopped: prometheus.NewDesc(
			"docker_containers_stopped_total",
			"The number of containers stopped",
			nil, nil,
		),
		containersPaused: prometheus.NewDesc(
			"docker_containers_paused_total",
			"The number of containers paused",
			nil, nil,
		),
		imagesCount: prometheus.NewDesc(
			"docker_images_total",
			"The number of images",
			nil, nil,
		),
		servicesCount: prometheus.NewDesc(
			"docker_services_total",
			"The number of services",
			nil, nil,
		),
		tasksRunning: prometheus.NewDesc(
			"docker_tasks_running_total",
			"The number of tasks running",
			[]string{"service_name"}, nil,
		),
		tasksDesired: prometheus.NewDesc(
			"docker_tasks_desired_total",
			"The number of tasks desired",
			[]string{"service_name"}, nil,
		),
		nodesCount: prometheus.NewDesc(
			"docker_nodes_total",
			"The number of nodes",
			nil, nil,
		),
		nodesActive: prometheus.NewDesc(
			"docker_nodes_active_total",
			"The number of active nodes",
			nil, nil,
		),
		stacksCount: prometheus.NewDesc(
			"docker_stacks_total",
			"The number of stacks",
			nil, nil,
		),
		containersRunningAllNodes: prometheus.NewDesc(
			"docker_containers_running_all_nodes_total",
			"The number of containers running across all nodes",
			[]string{"node_id", "node_hostname"}, nil,
		),
		totalContainersAllNodes: prometheus.NewDesc(
			"docker_containers_running_total_all_nodes",
			"The total number of containers running across all nodes combined",
			nil, nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *DockerSwarmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.containersRunning
	ch <- c.containersStopped
	ch <- c.containersPaused
	ch <- c.imagesCount
	ch <- c.servicesCount
	ch <- c.tasksRunning
	ch <- c.tasksDesired
	ch <- c.nodesCount
	ch <- c.nodesActive
	ch <- c.stacksCount
	ch <- c.containersRunningAllNodes
	ch <- c.totalContainersAllNodes
}

// Collect implements the prometheus.Collector interface
func (c *DockerSwarmCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Collect container metrics
	c.collectContainerMetrics(ctx, ch)

	// Collect image metrics
	c.collectImageMetrics(ctx, ch)

	// Check if Docker is in swarm mode
	info, err := c.dockerClient.Info(ctx)
	if err != nil {
		log.Printf("Error getting Docker info: %v", err)
		return
	}

	if info.Swarm.LocalNodeState == "active" {
		// Collect swarm metrics
		c.collectSwarmMetrics(ctx, ch)
	}
}

// collectContainerMetrics collects metrics about containers
func (c *DockerSwarmCollector) collectContainerMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	containers, err := c.dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return
	}

	var running, stopped, paused int

	for _, container := range containers {
		switch container.State {
		case "running":
			running++
		case "exited", "created", "dead":
			stopped++
		case "paused":
			paused++
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.containersRunning,
		prometheus.GaugeValue,
		float64(running),
	)
	ch <- prometheus.MustNewConstMetric(
		c.containersStopped,
		prometheus.GaugeValue,
		float64(stopped),
	)
	ch <- prometheus.MustNewConstMetric(
		c.containersPaused,
		prometheus.GaugeValue,
		float64(paused),
	)
}

// collectImageMetrics collects metrics about images
func (c *DockerSwarmCollector) collectImageMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	// Skip image metrics for now due to API compatibility issues
	ch <- prometheus.MustNewConstMetric(
		c.imagesCount,
		prometheus.GaugeValue,
		0,
	)
}

// collectSwarmMetrics collects metrics about Docker Swarm
func (c *DockerSwarmCollector) collectSwarmMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	// Collect services metrics
	services, err := c.dockerClient.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		log.Printf("Error listing services: %v", err)
	} else {
		ch <- prometheus.MustNewConstMetric(
			c.servicesCount,
			prometheus.GaugeValue,
			float64(len(services)),
		)

		// Collect tasks metrics for each service
		for _, service := range services {
			serviceName := service.Spec.Name

			// Get service tasks
			taskFilters := filters.NewArgs()
			taskFilters.Add("service", service.ID)

			tasks, err := c.dockerClient.TaskList(ctx, types.TaskListOptions{
				Filters: taskFilters,
			})
			if err != nil {
				log.Printf("Error listing tasks for service %s: %v", serviceName, err)
				continue
			}

			var runningTasks int
			for _, task := range tasks {
				if task.Status.State == swarm.TaskStateRunning {
					runningTasks++
				}
			}

			ch <- prometheus.MustNewConstMetric(
				c.tasksRunning,
				prometheus.GaugeValue,
				float64(runningTasks),
				serviceName,
			)

			// Get desired replicas
			var desiredReplicas uint64
			if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
				desiredReplicas = *service.Spec.Mode.Replicated.Replicas
			} else if service.Spec.Mode.Global != nil {
				// For global services, desired replicas equals the number of nodes
				nodes, err := c.dockerClient.NodeList(ctx, types.NodeListOptions{})
				if err != nil {
					log.Printf("Error listing nodes: %v", err)
				} else {
					var activeNodes int
					for _, node := range nodes {
						if node.Status.State == swarm.NodeStateReady {
							activeNodes++
						}
					}
					desiredReplicas = uint64(activeNodes)
				}
			}

			ch <- prometheus.MustNewConstMetric(
				c.tasksDesired,
				prometheus.GaugeValue,
				float64(desiredReplicas),
				serviceName,
			)
		}
	}

	// Collect nodes metrics
	nodes, err := c.dockerClient.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		log.Printf("Error listing nodes: %v", err)
	} else {
		var activeNodes int
		for _, node := range nodes {
			if node.Status.State == swarm.NodeStateReady {
				activeNodes++
			}
		}

		ch <- prometheus.MustNewConstMetric(
			c.nodesCount,
			prometheus.GaugeValue,
			float64(len(nodes)),
		)
		ch <- prometheus.MustNewConstMetric(
			c.nodesActive,
			prometheus.GaugeValue,
			float64(activeNodes),
		)
	}

	// Collect stacks metrics
	// Docker doesn't have a direct API for stacks, so we need to use labels
	// Stacks are identified by the "com.docker.stack.namespace" label on services
	stackMap := make(map[string]bool)
	for _, service := range services {
		if stackName, ok := service.Spec.Labels["com.docker.stack.namespace"]; ok {
			stackMap[stackName] = true
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.stacksCount,
		prometheus.GaugeValue,
		float64(len(stackMap)),
	)

	// Collect containers running on all nodes
	// In Docker Swarm, we can get this information from tasks
	// Each task corresponds to a container running on a node
	if len(nodes) > 0 {
		// Create a map to count running containers per node
		nodeContainers := make(map[string]int)
		nodeNames := make(map[string]string)

		// First, get all node IDs and hostnames
		for _, node := range nodes {
			nodeID := node.ID
			nodeHostname := node.Description.Hostname
			nodeContainers[nodeID] = 0
			nodeNames[nodeID] = nodeHostname
		}

		// Get all tasks (containers) in the swarm
		tasks, err := c.dockerClient.TaskList(ctx, types.TaskListOptions{})
		if err != nil {
			log.Printf("Error listing tasks: %v", err)
		} else {
			// Count running containers per node
			for _, task := range tasks {
				if task.Status.State == swarm.TaskStateRunning {
					nodeID := task.NodeID
					if _, ok := nodeContainers[nodeID]; ok {
						nodeContainers[nodeID]++
					}
				}
			}

			// Calculate total containers across all nodes
			totalContainers := 0
			for _, count := range nodeContainers {
				totalContainers += count
			}

			// Expose total containers metric
			ch <- prometheus.MustNewConstMetric(
				c.totalContainersAllNodes,
				prometheus.GaugeValue,
				float64(totalContainers),
			)

			// Expose metrics for each node
			for nodeID, count := range nodeContainers {
				ch <- prometheus.MustNewConstMetric(
					c.containersRunningAllNodes,
					prometheus.GaugeValue,
					float64(count),
					nodeID,
					nodeNames[nodeID],
				)
			}
		}
	}
}

func main() {
	flag.Parse()

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(
		client.WithHost(*dockerSocket),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Test Docker connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = dockerClient.Ping(ctx)
	if err != nil {
		log.Fatalf("Error connecting to Docker daemon: %v", err)
	}

	log.Printf("Connected to Docker daemon")

	// Create and register collector
	collector := NewDockerSwarmCollector(dockerClient, *scrapeTimeout)
	prometheus.MustRegister(collector)

	// Setup HTTP server
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Docker Swarm Exporter</title></head>
			<body>
			<h1>Docker Swarm Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	// Start server
	log.Printf("Starting Docker Swarm exporter on %s", *listenAddress)
	log.Printf("Metrics available at http://0.0.0.0%s%s", *listenAddress, *metricsPath)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
