#!/bin/bash
# 사용법을 출력하는 함수
usage() {
    echo "Usage: $0 {ingress|istio}"
    exit 1
}
# 인자가 제공되지 않은 경우 사용법을 출력합니다.
if [ "$#" -ne 1 ]; then
    usage
fi
# 스크립트가 위치한 디렉토리를 BASE_DIR로 사용
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 인자로 받은 값에 따라 CFG_FILE을 설정합니다.
case $1 in
    ingress)
        CFG_FILE="ingress-haproxy.cfg"
        ;;
    istio)
        CFG_FILE="istio-haproxy.cfg"
        ;;
    *)
        usage
        ;;
esac
# 현재 활성화된 설정 파일을 제거합니다.
sudo rm -f $BASE_DIR/haproxy.cfg
# 설정 파일을 교체합니다.
cp "$BASE_DIR/$CFG_FILE" "$BASE_DIR/haproxy.cfg"
# Docker Compose를 재시작합니다.
cd "$BASE_DIR"
docker-compose down
docker-compose up -d
