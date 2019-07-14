# Kubernetes completed pod cleaner


## Features

- Delete pod with pattern. <br/>
**You can specify pods with multiple patterns,** the format is `namespace/name`. For example if you want to delete pods only in the namespace `foo` the pattern should be `foo/*`.
- TTL (seconds). <br/>
It has TTL, only delete pods which has overed TTL seconds after completed.
- Delete the owner job together. <br/>
Delete the job which own the pod.
- It works in both **in-side** and **out-side** of cluster.
- Support docker image.

## Command Options
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

## Usage

### Out of cluster
After installation, build and run.
```bash
$ git clone git@github.com:hanjunlee/completed-pod-cleaner.git
$ cd completed-pod-cleaner
# go version >= 1.11
$ go build -o completed-pod-cleaner ./cmd
```
Note that you can specify the configuration path of kubernetes with `KUBECONFIG` environment. if you don't set `KUBECONFIG`, it will use `~/.kube/config` as the path.
```
$ completed-pod-cleaner -pattern 'foo/*' -pattern 'bar/*'  -ttl 3600 -job
```

### In cluster
Use the files in `./deploy` directory. Modify argument of command in the `./deploy/deployment.yaml` file.
```
$ kubectl apply -f ./deploy/service-account.yaml
$ kubectl apply -f ./deploy/deployment.yaml
```
