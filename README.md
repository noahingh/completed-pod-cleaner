# Kubernetes completed pod cleaner


## Features

- Delete pod with pattern. <br/>
**You can specify pods with multiple patterns,** the format is `namespace/name`. For example if you want to delete pods only in the namespace `foo` the pattern should be `foo/*`.
- TTL (seconds). <br/>
It has TTL, only delete pods which has overed TTL seconds after completed.
- Delete the owner job. <br/>
Delete the job which own the pod.

## Usage

### Command
```bash
completed-pod-cleaner -h
  -debug
    	Debug mode.
  -dry-run
    	Dry run mode, it does not delete pods.
  -job
    	Delete the job of pod together if the pod is ownered by the job.
  -kubeconfig string
    	Path to a kubeconfig. Only required if out-of-cluster.
  -pattern value
    	(list) The completed pods will be deleted when the pod match to pattern.
    	The format of pattern is "namespace/name" and you can use the wildcard(i.e '*').
  -thread int
    	The count of worker. (default 1)
  -ttl int
    	TTL seconds after the pod completed.
```

e.g)
```
$ completed-pod-cleaner -kubeconfig ~/.kube/config -pattern 'foo/*' -pattern 'bar/*'  -ttl 3600 -job
```

### Pod
See `./deploy` directory.
