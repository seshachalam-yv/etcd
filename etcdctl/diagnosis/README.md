# etcd-diagnosis

`etcd-diagnosis` collects a set of troubleshooting details from every endpoint of an etcd cluster. When the `--offline` flag is specified it analyzes an etcd data directory instead.

## Installation

Install the tool by running the following command from the etcd source directory:

```bash
$ go install ./tools/etcd-diagnosis
```

## Usage

```bash
$ etcd-diagnosis --help
```

### Examples

Run an online diagnosis against a running cluster:

```bash
$ etcd-diagnosis --endpoints=http://127.0.0.1:2379
```

Analyze data directory offline:

```bash
$ etcd-diagnosis --offline --data-dir /var/lib/etcd
```
