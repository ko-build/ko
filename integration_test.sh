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

pushd "$GOPATH/src/github.com/google/ko" || exit 1

echo "Building ko"

RESULT="$(GO111MODULE="on" GOFLAGS="-mod=vendor" go build .)"

echo "Beginning scenarios."

FILTER="[^ ]local[^ ]*"

echo "1. Go module auto mode should create an image that outputs 'Hello World' when run outside the module."

pushd .. || exit 1
RESULT="$(GO111MODULE=auto GOFLAGS="-mod=vendor" ./ko/ko build --local "$GOPATH/src/github.com/google/ko/test"  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

popd || exit 1

echo "2. Auto inside the module with vendoring should output Hello there"

RESULT="$(GO111MODULE=auto GOFLAGS="-mod=vendor" ./ko build --local "$GOPATH/src/github.com/google/ko/test"  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "3. Auto inside the module without vendoring should output Hello there"
RESULT="$(GO111MODULE=auto GOFLAGS="" ./ko build --local "$GOPATH/src/github.com/google/ko/test"  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "4. On inside the module with vendor should output Hello there."
RESULT="$(GO111MODULE=on GOFLAGS="-mod=vendor" ./ko build --local "$GOPATH/src/github.com/google/ko/test"  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "5. On inside the module without vendor should output Hello there"
RESULT="$(GO111MODULE=on GOFLAGS="" ./ko build --local "$GOPATH/src/github.com/google/ko/test"  | grep "$FILTER" | xargs -I% docker run %)"
if [[ "$RESULT" != *"Hello there"* ]]; then
  echo "Test FAILED. Saw $RESULT" && exit 1
else
  echo "Test PASSED"
fi

echo "6. On outside the module should fail."
pushd .. || exit 1
GO111MODULE=on ./ko/ko build --local "$GOPATH/src/github.com/google/ko/test" && exit 1
popd || exit 1

echo "7. On outside with build config specifying the test module builds."
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

echo "9. Linux capabilities."
pushd test/build-configs || exit 1
# run as non-root user with net_bind_service cap granted
docker_run_opts="--user 1 --cap-add=net_bind_service"
RESULT="$(GO111MODULE=on GOFLAGS="" ../../ko build --local ./caps/cmd | grep "$FILTER" | xargs -I% docker run $docker_run_opts %)"
if [[ "$RESULT" != "No capabilities" ]]; then
  echo "Test FAILED. Saw '$RESULT' but expected 'No capabilities'. Docker 'cap-add' must have no effect unless matching capabilities are granted to the file." && exit 1
fi
# build with a different config requesting net_bind_service file capability
RESULT_WITH_FILE_CAPS="$(KO_CONFIG_PATH=caps.ko.yaml GO111MODULE=on GOFLAGS="" ../../ko build --local ./caps/cmd | grep "$FILTER" | xargs -I% docker run $docker_run_opts %)"
if [[ "$RESULT_WITH_FILE_CAPS" !=  "Has capabilities"* ]]; then
  echo "Test FAILED. Saw '$RESULT_WITH_FILE_CAPS' but expected 'Has capabilities'. Docker 'cap-add' must work when matching capabilities are granted to the file." && exit 1
else
  echo "Test PASSED"
fi
popd || exit 1

popd || exit 1
popd || exit 1

export GOPATH="$ORIGINAL_GOPATH"
