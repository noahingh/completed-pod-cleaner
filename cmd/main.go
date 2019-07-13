package main

import (
	"flag"
	"time"

	"github.com/hanjunlee/completed-pod-cleaner/pkg/controller"
	log	"github.com/sirupsen/logrus"

	"k8s.io/sample-controller/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfigPath string
	debug bool
	threadiness int

	patterns Patterns
	ttl int
	job bool
	dryRun bool
)

func init() {
	flag.StringVar(&kubeconfigPath, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.BoolVar(&debug, "debug", false, "Debug mode.")
	flag.IntVar(&threadiness, "thread", 1, "The count of worker.")

	flag.Var(&patterns, "pattern", `(list) The completed pods will be deleted when the pod match to pattern.
The format of pattern is "namespace/name" and you can use the wildcard(i.e '*').`)
	flag.IntVar(&ttl, "ttl", 0, "TTL seconds after the pod completed.")
	flag.BoolVar(&job, "job", false, "Delete the job of pod together if the pod is ownered by the job.")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run mode, it does not delete pods.")
}


func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	stopCh := signals.SetupSignalHandler()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("can not get kube config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error building kubernetes clientset: %s", err.Error())
	}

	log.Infof("run the controller with options: patterns: %s, ttl: %d, job: %t, dry-run: %t", patterns, ttl, job, dryRun)
	factory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	controller := controller.NewController(
		clientset, 
		factory.Core().V1().Pods(),
		patterns,
		ttl,
		job,
		dryRun,
	)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh))
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	factory.Start(stopCh)

	if err = controller.Run(threadiness, stopCh); err != nil {
		log.Fatalf("error running controller: %s", err.Error())
	}
}