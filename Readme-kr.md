# OpenMSA
아래 이미지를 클릭하시면 YouTube 영상을 시청하실 수 있습니다.

<a href="https://youtu.be/EZhSwWLVGWg?si=TFXeazFOu5KoonOw">
  <img src="https://github.com/user-attachments/assets/a1ffb8f0-62b0-42be-ad33-2f2aecbe5116" alt="Example Image" width="400">
</a>

OpenMSA는 다중 마스터 구성 및 로드 밸런싱을 지원하는 다양한 Kubernetes 배포판(RKE2, Kubeadm, K3S)을 위한 올인원 클러스터 관리 솔루션으로, 클러스터 및 카탈로그 서비스의 배포를 자동화합니다.

## 목차
- [시스템 아키텍처](#시스템-아키텍처)
- [시스템 요구사항](#시스템-요구사항)
- [지원 인프라](#지원-인프라)
- [설치 프로세스](#설치-프로세스)
- [설치 후 구성](#설치-후-구성)
- [구성 상세](#구성-상세)

## 시스템 아키텍처

### 핵심 구성요소
- **Go 바이너리**: OpenMSA 실행 파일
- **GitHub Releases**: 바이너리 및 설치 스크립트 배포 플랫폼
- **설치 스크립트**: 시스템 의존성 자동 설정
- **Ansible 스크립트**: OS 및 Kubernetes 변종에 따른 설치 오케스트레이션
- **구성 디렉토리**: `/etc/openmsa`에 수정 가능한 Ansible 플레이북 포함

### 아키텍처 다이어그램
<img src="https://github.com/user-attachments/assets/ccb7f367-6335-4bf2-a995-4e14e26d9dd7" alt="Example Image" width="800">

## 시스템 요구사항

- 최소 3개 노드 필요
- 노드별 요구사항:
  - 메모리: 8GB 이상
  - CPU: 4코어 이상
  - 시스템 및 컨테이너 스토리지를 위한 충분한 디스크 공간

## 지원 인프라

### 검증된 운영체제
- AWS Amazon Linux 2
- Ubuntu 22.04 LTS
- Rocky Linux 8.9

### 노드 구성 규칙

| 노드 유형 | 역할 | 레이블 | 설명 |
|----------|------|--------|------|
| MGMT 노드 | control | control=true | 로드밸런서 및 Ansible Tower를 위한 MGMT(제어) 노드 |
| 마스터 노드 | control-plane | master=true | 마스터(컨트롤 플레인) 노드 |
| 워커 노드 | worker | worker=true | 워크로드 실행을 위한 워커 노드 |

### 지원 카탈로그 서비스

| 카탈로그명 | 기본값 | 노드-셀렉터 | 설명 |
|-----------|--------|-------------|------|
| ArgoCD | False | - | CI/CD 자동화 서버 |
| Jenkins | False | - | CI/CD 자동화 서버 |
| Gitlab | False | - | 소스 코드 관리 |
| Keycloak | False | - | 인증 관리 |
| Prometheus Stack | True | - | 모니터링 솔루션 |
| Opensearch | False | - | 검색 및 분석 |
| Opensearch Dashboard | False | - | 시각화 플랫폼 |
| Fluent-bit | False | - | 로그 프로세서 |
| Rancher Dashboard | True | - | Kubernetes 관리 UI |
| Jaeger | False | - | 분산 추적 |
| MySQL | False | db=true | 관계형 데이터베이스 |
| MariaDB | False | db=true | 관계형 데이터베이스 |
| PostgreSQL-HA | False | db=true | HA PostgreSQL 클러스터 |
| Redis-cluster | False | db=true | 인메모리 데이터 스토어 |
| Kafka | False | - | 이벤트 스트리밍 플랫폼 |
| Longhorn | True | storage=true | 분산 스토리지 |
| Minio | False | - | S3 호환 스토리지 |
| Istio | False | - | 서비스 메시 |
| Velero | False | - | 백업 및 마이그레이션 |
| Neuvector | False | - | 컨테이너 보안 |

### Kubernetes 지원 버전
- RKE2
- Kubeadm
- K3S

### 로드 밸런싱
- 분산 로드 관리를 위한 HAProxy 구현

## 설치 프로세스

### 사전 요구사항
- 패스워드 인증이 가능한 root 계정 접근
- 시스템 수정을 위한 root 권한
- 패키지 다운로드를 위한 인터넷 연결

### 빠른 시작

1. 설치 스크립트 다운로드 및 실행:
```bash
curl -L https://github.com/yuseok-jeong/openmsa-releases/releases/download/installer-v1.0.0/install.sh | sudo bash
```

2. OpenMSA 실행:
```bash
openmsa
```

### 설치 단계

사전 설치 프로세스 포함 사항:
1. OpenMSA 바이너리 배포
2. Ansible 구성
3. OpenSSH 설정 (Ansible 자동화 필수)
4. Python 설치 (Ansible 의존성)

### 시스템 구성
- 노드 구성 규칙에 따른 서버 정보 구성
<img src="https://github.com/user-attachments/assets/b958ffc8-0409-4cf4-886c-484560491fcd" alt="Example Image" width="150">

- Ansible 연결 설정
<img src="https://github.com/user-attachments/assets/e1e90524-ee7b-4d98-946f-5cafe4e85d25" alt="Example Image" width="250">

- Kubernetes 배포 유형 선택
<img src="https://github.com/user-attachments/assets/416b8e93-54e8-4661-bb7e-986cbfd7a890" alt="Example Image" width="250">

- 원하는 카탈로그 옵션 선택
<img src="https://github.com/user-attachments/assets/13cb2d7b-1d1c-4f42-8196-1a06aa54aa34" alt="Example Image" width="150">

### 클러스터 배포
- 시스템 구성 후 클러스터 및 카탈로그 서비스 배포
<img src="https://github.com/user-attachments/assets/321d9b36-f55b-49cd-937a-33d83c880216" alt="Example Image" width="250">

## 설치 후 구성

### 클러스터 검증

Kubernetes 클러스터 상태 확인:
```bash
kubectl get nodes
kubectl get pods -A
```

클러스터 관리를 위한 K9s 사용:
```bash
k9s
```

### 카탈로그 서비스 접근

1. hosts 파일 업데이트:
   - `/etc/hosts`에 MGMT 노드의 IP 주소 추가
2. 구성된 ingress 주소를 통해 카탈로그 UI 접근
3. 각 배포된 서비스의 운영 상태 확인

## 구성 상세

### 구성 관리
- 모든 Ansible 플레이북은 `/etc/openmsa` 디렉토리에 위치
- 특정 요구사항에 따라 플레이북 사용자 정의 가능

### 자동화된 작업
시스템이 자동으로 관리하는 항목:
- Ansible 설정 및 연결
- 운영체제별 구성
- Kubernetes 변종 배포
- 카탈로그 서비스 설치 및 구성

## 문제 해결
알려진 이슈
- 추후 업데이트 예정
