#!/bin/bash

# 레지스트리 검출 정규식
REGISTRY_REGEX_PATTERN='[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?'

# Helm 저장소 정보
HELM_REPOSITORY_URL=https://charts.gitlab.io/

# Helm 차트 명
HELM_CHART_NAME=gitlab

# Helm 차트 버전
HELM_CHART_VERSION=7.11.2

# Docker 저장소 정보
DOCKER_REGISTRY_IP=10.71.163.68:5000

DOWNLOAD_DIR=images
IMAGE_LIST=images.txt

helm repo add registry $HELM_REPOSITORY_URL

helm template versionfinder gitlab/$HELM_CHART_NAME \
    --set global.edition=ce \
    --set certmanager-issuer.email=example@example.com \
    --version $HELM_CHART_VERSION | grep 'image:' | tr -d '[[:blank:]]' | sort --unique | tr -d \" | sed 's/image://g' | awk -F '@' '{print $1}' > $IMAGE_LIST

GITLAB_VERSION=$(echo $(helm template versionfinder gitlab/$HELM_CHART_NAME \
                --set global.edition=ce \
                --set certmanager-issuer.email=example@example.com \
                --version $HELM_CHART_VERSION | grep 'gitlabVersion:' | tr -d '[[:blank:]]' | tr -d \" | sed 's/gitlabVersion://g'))

# Helm Command `helm template ~`가 아래의 이미지를 감지할 수 없으므로, 직접 추가 (이미지 누락 발생)
echo "registry.gitlab.com/gitlab-org/build/cng/cfssl-self-sign:v$GITLAB_VERSION" >> $IMAGE_LIST 

# GitLab Runner Base Image
echo "docker.io/docker:26.1.3" >> $IMAGE_LIST

# Sample Source Build Image (Java)
echo "docker.io/maven:3.9.2-eclipse-temurin-17" >> $IMAGE_LIST

# Sample Source Dockerize Image
echo "gcr.io/kaniko-project/executor:v1.14.0-debug" >> $IMAGE_LIST

while IFS= read -r IMAGE
do
    if [[ $IMAGE =~ $REGISTRY_REGEX_PATTERN ]]
    then
        TARGET_IMAGE=$(echo $IMAGE | awk -v REGISTRY=$DOCKER_REGISTRY_IP '{
            split($0, arr, "/")
            sub(arr[1],"", $0)
        } END {
            print REGISTRY$0
        }')
    else
        TARGET_IMAGE=$(echo $DOCKER_REGISTRY_IP/$IMAGE)
    fi
    
    docker pull $IMAGE

    docker tag $IMAGE $TARGET_IMAGE

    docker rmi $IMAGE

    echo -e "대상 이미지 : $TARGET_IMAGE"

    FILE_NAME=$(echo $"$(basename "$TARGET_IMAGE")" | awk -F ':' '{print $1}')

    mkdir -p $DOWNLOAD_DIR

    docker save -o $DOWNLOAD_DIR/$FILE_NAME.tar $TARGET_IMAGE

    echo -e "이미지 다운로드 성공 : $FILE_NAME.tar\\n"

    docker rmi $TARGET_IMAGE
done < $IMAGE_LIST