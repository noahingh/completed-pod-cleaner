package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hanjunlee/completed-pod-cleaner/pkg/controller"
	log "github.com/sirupsen/logrus"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/sample-controller/pkg/signals"
)

const (
	defaultThreadiness int = 1
)

var (
	debug       = flag.Bool("d", false, "")
	dryRun      = flag.Bool("dry-run", false, "")
	job         = flag.Bool("job", false, "")
	patterns    Patterns
	threadiness = flag.Int("t", defaultThreadiness, "")
	ttl         = flag.Int("ttl", 0, "")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprint(flag.CommandLine.Output(), `
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
	`)
	}

	flag.BoolVar(debug, "debug", false, "")
	flag.Var(&patterns, "p", "")
	flag.Var(&patterns, "pattern", "")
	flag.IntVar(threadiness, "thread", defaultThreadiness, "")
}

func main() {
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	stopCh := signals.SetupSignalHandler()

	var clientset kubernetes.Interface
	if _, err := rest.InClusterConfig(); err == nil {
		log.Info("it is running in the cluster, uses the service account kubernetes gives to pod.")
		clientset = GetClient()
	} else {
		log.Info("it is running out of the cluster, use the user's configuration.")
		clientset = GetClientOutOfCluster()
	}

	log.Infof("run the controller with options: patterns: %s, ttl: %d, job: %t, dry-run: %t", patterns, *ttl, *job, *dryRun)
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	controller := controller.NewController(
		clientset,
		factory.Core().V1().Pods(),
		patterns,
		*ttl,
		*job,
		*dryRun,
	)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh))
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	factory.Start(stopCh)

	if err := controller.Run(*threadiness, stopCh); err != nil {
		log.Fatalf("error running controller: %s", err.Error())
	}
}
