# OpenMSA
Click on the image below to watch the video on YouTube

<a href="https://youtu.be/EZhSwWLVGWg?si=TFXeazFOu5KoonOw">
  <img src="https://github.com/user-attachments/assets/a1ffb8f0-62b0-42be-ad33-2f2aecbe5116" alt="Example Image" width="400">
</a>

OpenMSA is an all-in-one cluster management solution that supports multiple Kubernetes distributions (RKE2, Kubeadm, K3S) with multi-master configuration and load balancing, while automating the deployment of clusters and catalog services.

## Table of Contents
- [System Architecture](#system-architecture)
- [System Requirements](#system-requirements)
- [Supported Infrastructure](#supported-infrastructure)
- [Installation Process](#installation-process)
- [Post-Installation](#post-installation)
- [Configuration Details](#configuration-details)

## System Architecture

### Core Components
- **Go Binary**: Main OpenMSA executable
- **GitHub Releases**: Distribution platform for binaries and installation scripts
- **Installation Script**: Automated setup for system dependencies
- **Ansible Scripts**: Orchestrates installation based on OS and Kubernetes variants
- **Configuration Directory**: `/etc/openmsa` contains modifiable Ansible playbooks

### Architecture Diagram
<img src="https://github.com/user-attachments/assets/ccb7f367-6335-4bf2-a995-4e14e26d9dd7" alt="Example Image" width="800">

## System Requirements

- Minimum of 3 nodes
- Per node requirements:
  - Memory: 8GB or more
  - CPU: 4 cores or more
  - Sufficient disk space for system and container storage

## Supported Infrastructure

### Validated Operating Systems
- AWS Amazon Linux 2
- Ubuntu 22.04 LTS
- Rocky Linux 8.9

### Node Configuration Rules


| Node Type     | Role          | Label         | Description |
|--------------|---------------|---------------|-------------|
| MGMT Node  | control        | control=true   | MGMT(control) node for Loadbalancer and ansible Tower |
| Master Node   | control-plane       | master=true  | Master(control plane) nodes |
| Worker Node   | worker | worker=true   | Worker nodes for workload execution  |



### Supported Catalog Services

| Catalog Name | Default | Node-Selector | Description |
|-------------|---------|---------------|-------------|
| ArgoCD | False | - | CI/CD automation server |
| Jenkins | False | - | CI/CD automation server |
| Gitlab | False | - | Source code management |
| Keycloak | False | - | Identity management |
| Prometheus Stack | True | - | Monitoring solution |
| Opensearch | False | - | Search and analytics |
| Opensearch Dashboard | False | - | Visualization platform |
| Fluent-bit | False | - | Log processor |
| Rancher Dashboard | True | - | Kubernetes management UI |
| Jaeger | False | - | Distributed tracing |
| MySQL | False | db=true | Relational database |
| MariaDB | False | db=true | Relational database |
| PostgreSQL-HA | False | db=true | HA PostgreSQL cluster |
| Redis-cluster | False | db=true | In-memory data store |
| Kafka | False | - | Event streaming platform |
| Longhorn | True | storage=true | Distributed storage |
| Minio | False | - | S3-compatible storage |
| Istio | False | - | Service mesh |
| Velero | False | - | Backup and migration |
| Neuvector | False | - | Container security |


### Kubernetes Variants
- RKE2
- Kubeadm
- K3S

### Load Balancing
- HAProxy implementation for distributed load management

## Installation Process

### Prerequisites
- Root account access with password authentication
- System modifications require root privileges
- Internet connectivity for package downloads

### Quick Start

1. Download and execute the installation script:
```bash
curl -L https://github.com/yuseok-jeong/openmsa-releases/releases/download/installer-v1.0.0/install.sh | sudo bash
```

2. Launch OpenMSA:
```bash
openmsa
```

### Installation Steps

The pre-installation process includes:
1. OpenMSA binary deployment
2. Ansible configuration
3. OpenSSH setup (required for Ansible automation)
4. Python installation (Ansible dependency)

### System Configuration
- Configure server information based on Node Configuration Rules
 <img src="https://github.com/user-attachments/assets/b958ffc8-0409-4cf4-886c-484560491fcd" alt="Example Image" width="150">


- Set up Ansible connectivity
 <img src="https://github.com/user-attachments/assets/e1e90524-ee7b-4d98-946f-5cafe4e85d25" alt="Example Image" width="250">


- Select Kubernetes deployment type

 <img src="https://github.com/user-attachments/assets/416b8e93-54e8-4661-bb7e-986cbfd7a890" alt="Example Image" width="250">


- Choose desired catalog options
<img src="https://github.com/user-attachments/assets/13cb2d7b-1d1c-4f42-8196-1a06aa54aa34" alt="Example Image" width="150">

### Deploy Cluster
- Deploy Your cluster and catalog services after system configuration
<img src="https://github.com/user-attachments/assets/321d9b36-f55b-49cd-937a-33d83c880216" alt="Example Image" width="250">

## Post-Installation

### Cluster Verification

Check Kubernetes cluster status:
```bash
kubectl get nodes
kubectl get pods -A
```

Use K9s for cluster management:
```bash
k9s
```

### Accessing Catalog Services

1. Update hosts file:
   - Edit `/etc/hosts` to include MGMT node's IP address
2. Access catalog UIs through configured ingress addresses
3. Verify each deployed service is operational

## Configuration Details

### Configuration Management
- All Ansible playbooks are located in `/etc/openmsa` directory
- Playbooks can be customized based on specific requirements

### Automated Operations
The system automatically manages:
- Ansible Setup and Connection
- Operating system-specific configurations
- Kubernetes variant deployment
- Catalog service installation and configuration

## Troubleshooting
Known Issues 
- To-Be Updated
