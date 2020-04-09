## node-labels-to-files

[![Build Status](https://api.travis-ci.org/bjhaid/node-labels-to-files.svg?branch=master)](https://travis-ci.org/bjhaid/node-labels-to-files)
[![Docker Pulls](https://img.shields.io/docker/pulls/bjhaid/node-labels-to-files.svg)](https://github.com/bjhaid/node-labels-to-files)

This is mainly [Tim Hockin's
idea](https://docs.google.com/document/d/1fH64mEfZH597luup-ZBfBNkiTVjfoFYGEa-G1G_TM6A/edit#heading=h.1fewofmjczp2), I just implemented it.

This addresses the problems identified in
https://github.com/kubernetes/kubernetes/issues/40610 but outside of
Kubernetes.

It retrieves Kubernetes node labels and writes them to some pre-defined
directory path.

It runs in 2 modes:
- Once: In this mode it retrieves the node labels sets the keys as filenames
in the pre-defined directory and the values as content of the files then
exits. In Kubernetes you use this mode if you were to run node-labels-to-files
as an init container.
- Always: In this mode it establishes a watch on the node and continuously
updates files in the pre-defined directory to match the keys of the node
labels. In Kubernetes you use this mode if you were to run
node-labels-to-files as a sidecar.

By default it will delete files that are stale or that it does not know about,
to prevent this please make sure you have the `-delete-stale-files` flag set
to false

### Usage

```bash
Usage of ./node-labels-to-files:
  -alsologtostderr
        log to standard error as well as files
  -delete-stale-files
        This determines if node-labels-to-path will delete stale files or files it is not aware of or keep them, by default it will delete them. Can be overriden via the env
ironment variable DELETE_STALE_FILES (default true)
  -directory string
        Directory to write the node labels in, if the directory does not exist node-labels-to-files will create it. Can be overridden via the environment variable DIRECTORY
  -kubeconfig string
        (optional) absolute path to the kubeconfig file. Can be overridden via the environment variable KUBECONFIG (default "/home/bjhaid/.kube/config")
  -log_backtrace_at value
        when logging hits line file:N, emit a stack trace
  -log_dir string
        If non-empty, write log files in this directory
  -log_file string
        If non-empty, use this log file
  -log_file_max_size uint
        Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
  -logtostderr
        log to standard error instead of files (default true)
  -mode string
        This determines the mode n works in, when it is set to once it retrieves the node labels and exits, if set to always it creates a watch on the node and will detect a
nd update the directory to reflect the labels when they change on the node. Acceptable options is either of always|onceCan be overriden via the environment variable MODE (de
fault "always")
  -nodename string
        Name of node whose label n should retrieve. Can be overridden via the environmentvariable NODENAME
  -skip_headers
        If true, avoid header prefixes in the log messages
  -skip_log_headers
        If true, avoid headers when opening log files
  -stderrthreshold value
        logs at or above this threshold go to stderr (default 2)
  -v value
        number for the log level verbosity
  -vmodule value
        comma-separated list of pattern=N settings for file-filtered logging
```

### Example:

```bash
$> ./node-labels-to-files \
  -mode=once \
  -delete-stale-files=false \
  -nodename=ip-my-awesome-node.ec2.internal \
  -directory=${PWD}/foo

tree ${PWD}/foo
foo
├── beta.kubernetes.io
│   ├── arch
│   ├── instance-type
│   └── os
├── failure-domain.beta.kubernetes.io
│   ├── region
│   └── zone
├── kops.k8s.io
│   └── instancegroup
├── kubernetes.io
│   ├── hostname
│   └── role
└── node-role.kubernetes.io
    └── node

    5 directories, 9 files
```

or

```bash
$> DIRECTORY=/tmp/foo NODENAME=ip-my-awesome-node.ec2.internal MODE=once ./node-labels-to-files
I0721 11:50:58.105222   24181 main.go:262] Starting node-labels-to-files
I0721 11:50:58.304381   24181 main.go:184] Creating sub-directory: /tmp/foo/kops.k8s.io
I0721 11:50:58.304535   24181 main.go:184] Creating sub-directory: /tmp/foo/node-role.kubernetes.io
I0721 11:50:58.304588   24181 main.go:184] Creating sub-directory: /tmp/foo/failure-domain.beta.kubernetes.io
I0721 11:50:58.304668   24181 main.go:184] Creating sub-directory: /tmp/foo/beta.kubernetes.io
I0721 11:50:58.304725   24181 main.go:184] Creating sub-directory: /tmp/foo/kubernetes.io
$> tree /tmp/foo
/tmp/foo
├── beta.kubernetes.io
│   ├── arch
│   ├── instance-type
│   └── os
├── failure-domain.beta.kubernetes.io
│   ├── region
│   └── zone
├── kops.k8s.io
│   └── instancegroup
├── kubernetes.io
│   ├── hostname
│   └── role
└── node-role.kubernetes.io
    └── node
```

Running in docker:

```bash
docker run \
  --rm \
  -v /tmp/foo:/tmp/foo \
  -v ~/.kube/config:/etc/kubeconfig \
  -e KUBECONFIG=/etc/kubeconfig \
  -e DIRECTORY=/tmp/foo \
  -e NODENAME=ip-my-awesome-node.ec2.internal \
  -e MODE=once \
  bjhaid/node-labels-to-files:v0.0.4 \
  /usr/bin/node-labels-to-files
I0721 17:10:57.688212       1 main.go:262] Starting node-labels-to-files
I0721 17:10:58.219404       1 main.go:184] Creating sub-directory: /tmp/foo/failure-domain.beta.kubernetes.io
I0721 17:10:58.220011       1 main.go:184] Creating sub-directory: /tmp/foo/beta.kubernetes.io
I0721 17:10:58.220641       1 main.go:184] Creating sub-directory: /tmp/foo/kops.k8s.io
I0721 17:10:58.221016       1 main.go:184] Creating sub-directory: /tmp/foo/kubernetes.io
I0721 17:10:58.221551       1 main.go:184] Creating sub-directory: /tmp/foo/node-role.kubernetes.io
```

Note in Kubernetes you will not need to set the KUBECONFIG environment
variable.


### Kubernetes Example

[k8s-example.yaml](examples/k8s-example.yaml)

### Contributing

Make your changes, run:

```bash
docker build .
```

If the image builds successfully open a PR.
