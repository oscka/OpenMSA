#!/bin/bash

# 레지스트리 검출 정규식
REGISTRY_REGEX_PATTERN='[^/]+\.[^/.]+/([^/.]+/)?[^/.]+(:.+)?'

DOCKER_REGISTRY_IP=10.71.163.68:5000

DOWNLOAD_DIR=images
IMAGE_LIST=images.txt

for ENTRY in "$DOWNLOAD_DIR"/*
do
    sudo /var/lib/rancher/rke2/bin/ctr --address /run/k3s/containerd/containerd.sock --namespace k8s.io i import $ENTRY
done

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

    sudo /var/lib/rancher/rke2/bin/ctr --address /run/k3s/containerd/containerd.sock --namespace k8s.io i push $TARGET_IMAGE

    sudo /var/lib/rancher/rke2/bin/ctr --address /run/k3s/containerd/containerd.sock --namespace k8s.io i remove $TARGET_IMAGE
done < $IMAGE_LIST