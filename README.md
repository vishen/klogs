# Klogs
Yet another kubernetes log tools. `klogs` provides structured log searching for kubernetes container logs; exact and regex (golang stdlib) key value matching.

`klogs` is a kubernetes integration wrapper around [slearch](https://github.com/vishen/go-slearch), which provides structured log searching. It does a best effort to determine what format the current log line in; you can force it to only look for a certain log format with the `-t` command line argument.

`klogs` use your default kubernetes configuration to connect to the cluster, but it can be changed using the standard kubernetes config arguments `--kubeconfig` and `--kubecontext`

`klogs` will add the following key value fields to your logging line in the same log format, and each of these fields can be searched on as if they were in your log line:

* `namespace=<pod_namespace>`
* `pod_name=<pod_name>`
* `name=<container_name>`

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
  klogs [flags]

Flags:
  -c, --containers strings     kubernetes selector (label query) to filter on
  -h, --help                   help for klogs
  -d, --key_delimiter string   the string to split the key on for complex key queries
      --kubeconfig string      Path to kubernetes config
      --kubecontext string     Kubernetes context to use
  -m, --match strings          key and value to match on. eg: label1=value1
  -n, --namespace string       the kubernetes namespace to filter on
  -p, --print_keys strings     keys to print if a match is found
  -r, --regexp strings         key and value to regex match on. eg: label1=value*
  -s, --search_type string     the search type to use: 'and' or 'or' (default "and")
  -l, --selector strings       kubernetes selector (label query) to filter on. eg: app=api
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

## TODO
```
- command line argument for key existence
- command line argument for key not exists?
- add resource name arguments similar to kubectl
- add cmd arg for tailing or not
- watch for new pods
- option to not close streams gracefully?
```
