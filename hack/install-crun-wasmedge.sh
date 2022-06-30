#!/bin/bash

echo -e "Installing WasmEdge"
if [ -f install.sh ]; then
  rm -rf install.sh
fi
curl -L -o install.sh -q https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh
sudo chmod a+x install.sh
# use 0.9.1 because 0.10.0 had a breaking change which is not yet fixed in crun.
# See https://github.com/containers/crun/pull/933
sudo bash ./install.sh --path="/usr/local" -v 0.9.1
rm -rf install.sh

echo -e "Building and installing crun"
sudo apt install -y make git gcc build-essential pkgconf libtool libsystemd-dev \
    libprotobuf-c-dev libcap-dev libseccomp-dev libyajl-dev \
    go-md2man libtool autoconf python3 automake

git clone https://github.com/containers/crun /tmp/crun || true
cd /tmp/crun
./autogen.sh
./configure --with-wasmedge
make
