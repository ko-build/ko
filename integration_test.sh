#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "$0")

pushd "$ROOT_DIR"

ROOT_DIR="$(pwd)"

echo "Moving GOPATH into /tmp/ to test modules behavior."
export GOPATH="${GOPATH:-$(go env GOPATH)}"
export ORIGINAL_GOPATH="$GOPATH"
GOPATH="$(mktemp -d)"
export GOPATH

export CGO_ENABLED=0

GOARCH="${GOARCH:-$(go env GOARCH)}"

pushd "$GOPATH" || exit 1

echo "Copying ko to temp gopath."
mkdir -p "$GOPATH/src/github.com/google/ko"
cp -r "$ROOT_DIR/"* "$GOPATH/src/github.com/google/ko/"

pushd "$GOPATH/src/github.com/google/ko" || exit 1

echo "Building ko"

RESULT="$(go build .)"

echo "Beginning scenarios."

FILTER="[^ ]local[^ ]*"

echo "1. Test should create an image that outputs 'Hello World'."
RESULT="$(./ko build --local --platform="linux/$GOARCH" "$GOPATH/src/github.com/google/ko/test" | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "2. Test knative 'KO_FLAGS' variable is ignored."
# https://github.com/ko-build/ko/issues/1317
RESULT="$(KO_FLAGS="--platform=badvalue" ./ko build --local --platform="linux/$GOARCH" "$GOPATH/src/github.com/google/ko/test" | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "3. Linux capabilities."
pushd test/build-configs || exit 1
# run as non-root user with net_bind_service cap granted
docker_run_opts="--user 1 --cap-add=net_bind_service"
RESULT="$(../../ko build --local --platform="linux/$GOARCH" ./caps/cmd | grep "$FILTER" | xargs -I% docker run $docker_run_opts %)"
if [[ "$RESULT" != "No capabilities" ]]; then
  echo "Test FAILED. Saw '$RESULT' but expected 'No capabilities'. Docker 'cap-add' must have no effect unless matching capabilities are granted to the file." && exit 1
fi
# build with a different config requesting net_bind_service file capability
RESULT_WITH_FILE_CAPS="$(KO_CONFIG_PATH=caps.ko.yaml ../../ko build --local --platform="linux/$GOARCH"  ./caps/cmd | grep "$FILTER" | xargs -I% docker run $docker_run_opts %)"
if [[ "$RESULT_WITH_FILE_CAPS" !=  "Has capabilities"* ]]; then
  echo "Test FAILED. Saw '$RESULT_WITH_FILE_CAPS' but expected 'Has capabilities'. Docker 'cap-add' must work when matching capabilities are granted to the file." && exit 1
else
  echo "Test PASSED"
fi
popd || exit 1

popd || exit 1
popd || exit 1

export GOPATH="$ORIGINAL_GOPATH"
