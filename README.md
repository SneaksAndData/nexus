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
Nexus can be deployed in Single Cluster or Multi Cluster modes. Single Cluster mode consists of three components: at least one `scheduler`, a `maintainer` and at least one `receiver`, all deployed to a single Kubernetes cluster. Multi Cluster mode consists of:
- **Controller Cluster**, which has at least one `scheduler` and a `maintainer`
- One or more **Shard Clusters**, with at least one `receiver`. `scheduler` can also be deployed to these clusters in case an algorithm uses one of Nexus SDK's to create execution trees

### Maintainer
Maintainer is responsible for handling requests that are sitting in a queue more than expected, as well as requests that could not be converted to a Kubernetes Job for any reason, and for garbage collecting failed submissions.
Nexus stores request metadata and algorithm lifecycle-related information in a so-called checkpoint store. Maintainer scans that table on regular intervals, selects requests that are experiencing problems, i.e. they are stuck
in buffered state and tries to resolve the situation by either terminating the request, or sending it ahead of the main queue. In addition, Maintainer is responsible for watching for events emitted by algorithm pods and taking action in certain situations (OOMKill, ImagePullBackoff etc.)

### Scheduler

Scheduler is what makes it possible to run algorithms through Nexus. Each scheduler has a public API that can be used to submit runs and retrieve results. Moreover, each scheduler holds a separate virtual queue that it uses to process incoming requests.
Nexus relies on load balancer using round-robin algorithm when distributing requests between scheduler pods, so a horizontal autoscaler can be used to the maximum efficiency.

## Usage

-- TBD --

### Versioning

Nexus's API is versioned. Requests against a specific version with have URI suffix like this: `/algorithm/v1.0/run`. If a version is not specified, `latest` will be targeted - which includes experimental
and unstable features. For production use cases, always use one of the current production versions: 
- `/algorithm/v1.2/run`

### API Management
Adding new API paths must be reflected in Swagger docs, even though the app doesn't serve Swagger. Update the generated docs:
```shell
./swag init --parseDependency --parseInternal -g main.go
```

This is required for the API clients (Go and Python) to be updated correctly.