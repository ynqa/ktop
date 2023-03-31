# ktop

A visualized monitoring dashboard for Kubernetes.

`kubectl` also has `top` subcommand, but it is not able to:

- Watch usages regularly for Pod/Node
- Compare the usage of Pod resources with Node or limits/requests

`ktop` resolves these problems and has a more graphical dashboard.

## Installation

For MacOS:

```bash
$ brew tap ynqa/tap
$ brew install ktop
```

From source codes:

```bash
$ go get -u github.com/ynqa/ktop
```

## Usage
```
Kubernetes monitoring dashboard on terminal

Usage:
  ktop [flags]

Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default HTTP cache directory (default "/Users/ynqa/.kube/http-cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
  -C, --container-query string         container query (default ".*")
      --context string                 The name of the kubeconfig context to use
  -h, --help                           help for ktop
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -i, --interval duration              set interval (default 1s)
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
  -N, --node-query string              node query (default ".*")
  -P, --pod-query string               pod query (default ".*")
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```
