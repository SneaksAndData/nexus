# Introduction
Nexus is a lightweight, Kubernetes-native, high-throughput client-server proxy for machine learning, optimization and AI algorithms. Nexus allows data science teams to author and deploy their products to live environments with a few YAML manifests, and provides a simple flexible HTTP API for interacting with machine learning applications.

Core Nexus features include:
- Elastic virtual queues for incoming requests
    - Incoming request rate normalization via throttling to limits acceptable by Kubernetes API Server
    - Automatic scheduling of a delayed request ahead of a queue
- Allocation to autoscaling machine groups using Kubernetes affinity/toleration settings
- Plug-n-play algorithm execution and deployment via Kubernetes Custom Resources
- Multi-cluster support without extra configuration with the help of [Nexus Configuration Controller](https://github.com/SneaksAndData/nexus-configuration-controller)
- Input buffering and push down to algorithm container using secure URIs
- Processing rate from receiving to result on a scale of thousands of completions per second, depending algorithm execution time.
- HTTP API for launching runs and receiving results via presigned URLs.

 ## Design
Nexus can be deployed in Single Cluster or Multi Cluster modes. Single Cluster mode consists of three components: at least one `scheduler`, a `supervisor` and at least one `receiver`, all deployed to a single Kubernetes cluster. Multi Cluster mode consists of:
- **Controller Cluster**, which has at least one `scheduler`
- One or more **Shard Clusters**, with at least one `receiver` and a `supervisor`.
  - `scheduler` can also be deployed to these clusters in case an algorithm uses Nexus SDK to create execution trees

### Supervisor
Supervisor handles the following scenarios:
- Requests that were delayed due to a `scheduler` instance shutdown - those are picked up by the supervisor and submitted to the target cluster. 
- Misconfigured requests that could not be converted to a Kubernetes Job for any reason
- Garbage collecting failed submissions
- State accounting and garbage collecting submissions with container launch issues such as `ImagePullBackoff` or runtime failures such as `OOMKill` etc.

### Scheduler

Scheduler is what makes it possible to run algorithms through Nexus. Each scheduler has a public API that can be used to submit runs and retrieve results. Moreover, each scheduler holds a separate virtual queue that it uses to process incoming requests.
Nexus relies on load balancer using round-robin algorithm when distributing requests between scheduler pods, so a horizontal autoscaler should be used for production deployments.

## Usage



### Versioning

Nexus's API is versioned. Requests against a specific version with have URI suffix like this: `/algorithm/v1/run`. If a version is not specified, `latest` API will be targeted - which might include experimental
and unstable features. For production, always use a stable API version `/algorithm/v1/run`. Most changes tested under `latest` will eventually be integrated into `v1`. When a next major release `v2` comes along, `v1` will be a supported release until `v3` is released. Feature requests must be tagged by an API version they target - currently, `v1` only.


### API Management
Adding new API paths must be reflected in Swagger docs, even though the app doesn't serve Swagger. Update the generated docs:
```shell
./swag init --parseDependency --parseInternal -g main.go
```

This is required for the API clients (Go and Python) to be updated correctly. Note that until Swag 2.0 is released OpenAPI v3 model must be updated using [Swagger converter](https://converter.swagger.io/#/Converter/convertByContent)