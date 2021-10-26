#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(dirname "$0")

pushd "$ROOT_DIR"

ROOT_DIR="$(pwd)"

echo "Moving GOPATH into /tmp/ to test modules behavior."
export ORIGINAL_GOPATH="$GOPATH"
GOPATH="$(mktemp -d)"
export GOPATH

pushd "$GOPATH" || exit 1

echo "Copying ko to temp gopath."
mkdir -p "$GOPATH/src/github.com/google/ko"
cp -r "$ROOT_DIR/"* "$GOPATH/src/github.com/google/ko/"

echo "Downloading github.com/go-training/helloworld"
GO111MODULE=off go get -d github.com/go-training/helloworld

pushd "$GOPATH/src/github.com/google/ko" || exit 1

echo "Replacing hello world in vendor with TEST."
sed -i 's/Hello World/TEST/g' ./vendor/github.com/go-training/helloworld/main.go

echo "Building ko"

RESULT="$(GO111MODULE="on" GOFLAGS="-mod=vendor" go build .)"

echo "Beginning scenarios."

FILTER="[^ ]local[^ ]*"

echo "1. GOPATH mode should always create an image that outputs 'Hello World'"
RESULT="$(GO111MODULE=off ./ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello World"** ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "2. Go module auto mode should create an image that outputs 'Hello World' when run outside the module."

pushd .. || exit 1
RESULT="$(GO111MODULE=auto GOFLAGS="-mod=vendor" ./ko/ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello World"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

popd || exit 1

echo "3. Auto inside the module with vendoring should output TEST"

RESULT="$(GO111MODULE=auto GOFLAGS="-mod=vendor" ./ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"TEST"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "4. Auto inside the module without vendoring should output TEST"
RESULT="$(GO111MODULE=auto GOFLAGS="" ./ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"TEST"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "5. On inside the module with vendor should output TEST."
RESULT="$(GO111MODULE=on GOFLAGS="-mod=vendor" ./ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"TEST"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "6. On inside the module without vendor should output TEST"
RESULT="$(GO111MODULE=on GOFLAGS="" ./ko build --local github.com/go-training/helloworld  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"TEST"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "7. On outside the module should fail."
pushd .. || exit 1
GO111MODULE=on ./ko/ko build --local github.com/go-training/helloworld && exit 1
popd || exit 1

echo "8. On outside with build config specifying the test module builds."
pushd test/build-configs || exit 1
for app in foo bar ; do
  # test both local and fully qualified import paths
  for prefix in example.com . ; do
    import_path=$prefix/$app/cmd
    RESULT="$(GO111MODULE=on GOFLAGS="" ../../ko build --local $import_path | grep "$FILTER" | xargs -I% docker run %)"
    if [[ "$RESULT" != *"$app"* ]]; then
      echo "Test FAILED for $import_path. Saw $RESULT but expected $app" && exit 1
    else
      echo "Test PASSED for $import_path"
    fi
  done
done
popd || exit 1

popd || exit 1
popd || exit 1

export GOPATH="$ORIGINAL_GOPATH"
