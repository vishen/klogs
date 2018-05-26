# Klogs
Yet another kubernetes log tools. `klogs` provides structured log searching for kubernetes container logs; exact and regex (golang stdlib) key value matching.

`klogs` is a (mostly) drop-in replacement for `kubectl logs` except with a few differences;

* Defaults to logging all pods and containers in all namespaces
* Gives structured log searching like; exact and regex key value matching on `JSON` and `text` structured logs
* Will log additional information about the pod at the start of each the log line; `namespace`, `pod_name`, `container_name`
* When in follow mode, it will watch for new pods and automatically stream their logs

`klogs` use your default kubernetes configuration to connect to the cluster, but it can be changed using the standard kubernetes config arguments `--kubeconfig` and `--kubecontext`

Currently the only supported log format types are:

* JSON
* text

## Installing
```
$ go get -u github.com/vishen/klogs
$ GOPATH/bin/klogs
```

## Intalling from source
```
$ go build -o klogs .
$ ./klogs
```

## klogs command
```
Read stuctured logs from Kubernetes and filter out lines based on exact or regex matches. Currently only supports JSON and text logs.

Usage:
  klogs [-f] (POD) [-c CONTAINER] [flags]

Flags:
  -c, --containers strings     kubernetes selector (label query) to filter on
  -f, --follow                 tail the logs
  -h, --help                   help for klogs
  -d, --key_delimiter string   the string to split the key on for complex key queries
  -e, --key_exists strings     print lines that have these keys
      --kubeconfig string      Path to kubernetes config
      --kubecontext string     Kubernetes context to use
  -m, --match strings          key and value to match on. eg: label1=value1
  -n, --namespace string       the kubernetes namespace to filter on
  -p, --print_keys strings     keys to print if a match is found
  -r, --regexp strings         key and value to regex match on. eg: label1=value*
  -s, --search_type string     the search type to use: 'and' or 'or' (default "and")
  -l, --selector string        kubernetes selector (label query) to filter on. eg: app=api
  -t, --type string            the log type to use: 'json' or 'text'. If unspecified it will attempt to use all log types
  -v, --verbose                verbose output

```

## Examples
NOTE: The following examples work with both 'JSON' and 'text' log formats

```
# Searching for exact matches
$ klogs -m correlation-id=123123
severity="info" correlation-id="123123" msg="starting" user_id="7"
severity="info" correlation-id="123123" msg="processing" user_id="7"
severity="info" correlation-id="123123" msg="ending" user_id="7"

# Searching for exact multiple matches
$ klogs -m correlation-id=123123 -m user_id=7
severity="info" correlation-id="123123" msg="starting" user_id="7"
severity="info" correlation-id="123123" msg="processing" user_id="7"
severity="info" correlation-id="123123" msg="ending" user_id="7"

# Searching for regex matches
$ klogs -r correlation-id=123
severity="info" correlation-id="123123" msg="starting" user_id="7"
severity="info" correlation-id="123123" msg="processing" user_id="7"
severity="info" correlation-id="123123" msg="ending" user_id="7"

# Printing only certain keys with a match
$ klogs -m correlation-id=123123 -p msg,user_id
msg="starting" user_id="7"
msg="processing" user_id="7"
msg="ending" user_id="7"

# Printing only certain keys
$ klogs -m -p msg,user_id
msg="starting" user_id="7"
msg="processing" user_id="7"
msg="ending" user_id="7"
```
