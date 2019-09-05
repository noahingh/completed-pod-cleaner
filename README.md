# Kubernetes completed pod cleaner
The controller delete the completed pods which matched with patterns.

## Installation

### K8S
`kubectl apply -f https://raw.githubusercontent.com/hanjunlee/completed-pod-cleaner/master/deploy/kubernetes.yaml`.

### Local
```
git clone git@github.com:hanjunlee/completed-pod-cleaner.git
cd ./completed-pod-cleaner
go build -o completed-pod-cleaner ./cmd
```

% Note that you can specify the configuration path of kubernetes with `KUBECONFIG` environment. if you don't set `KUBECONFIG`, it will use `~/.kube/config` as the path.

## Features

- Delete pod with pattern. <br/>
**You can specify pods with multiple patterns,** the format is `namespace/name`. For example if you want to delete pods only in the namespace `foo` the pattern should be `foo/*`.
- TTL (seconds). <br/>
It has TTL, only delete pods which has overed TTL seconds after completed.
- Delete the job together if the pod is owned by the job. 
- It works in both **in-side** and **out-side** of cluster.

## Command
```bash
  -d, -debug
		Debug mode.
  -dry-run
		Dry run mode, it does not delete pods.
  -job
		Delete the job together if the pod is owned by the job.
  -p, -pattern
		(list) Match the pattern with the key of pod, the key is "namespace/name",
		and if matched the worker delete the completed pod. The pattern support wildcard("*").
  -t, -thread
		The count of worker, delete pods and jobs (default 1)
  -ttl
		TTL seconds after the pod completed (default 0).
```

## Example

### Pattern 
Suppose you want to delete completed pods in namespace `foo` and `bar`, you should run `completed-pod-cleaner -p foo/* -p bar/*`.

### Clean owner job
If you want to delete pod and also the job own the pod, you should run `completed-pod -p foo/* -job`
