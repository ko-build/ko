# Minhash LSH in Golang

[![Build Status](https://travis-ci.org/ekzhu/minhash-lsh.svg?branch=master)](https://travis-ci.org/ekzhu/minhash-lsh)
[![GoDoc](https://godoc.org/github.com/ekzhu/minhash-lsh?status.svg)](https://godoc.org/github.com/ekzhu/minhash-lsh)

[Documentation](https://godoc.org/github.com/ekzhu/minhash-lsh)

Install: `go get github.com/ekzhu/minhash-lsh`

## Run Benchmark

### Set file format

1. One set per line
2. Each set, all items are separated by whitespaces
3. If the parameter firstItemIsID is set to true,
   the first itme is the unique ID of the set.
4. The rest of the items with the following format: `<value>____<frequency>`

   * value is an unique element of the set
   * frequency is an integer count of the occurance of value
   * `____` (4 underscores) is the separator

### All Pair Benchmark

```
minhash-lsh-all-pair -input <set file name>
```
