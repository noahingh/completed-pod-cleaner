package controller

import (
	"fmt"
	"time"

	log	"github.com/sirupsen/logrus"
	"github.com/minio/minio/pkg/wildcard"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/apimachinery/pkg/util/wait"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// Controller delete completed pod if the key(namespace/name) are equal to the pattern
type Controller struct {
	log *log.Entry
	patterns []string
	ttl int
	job bool
	dryRun bool

	clientset kubernetes.Interface
	lister listers.PodLister
	synced cache.InformerSynced	
	workqueue workqueue.RateLimitingInterface
}

// NewController create the controller which delete the completed pod match to patterns.
func NewController(clientset kubernetes.Interface, informer informers.PodInformer, patterns []string, ttl int, job bool, dryRun bool) *Controller{
	c := &Controller{
		log: log.WithFields(log.Fields{"role": "controller"}),
		patterns: patterns,
		ttl: ttl,
		job: job,
		dryRun: dryRun,

		clientset: clientset,
		lister: informer.Lister(),
		synced: informer.Informer().HasSynced,
		workqueue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	
	c.log.Debug("add the events")
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.enqueuePod(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			c.enqueuePod(new)
		},
	})
	return c
}

// Run wait for syncronizing cache and run worker as many as threadiness.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.log.Debug("starting the controller")
	c.log.Info("waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.log.Info("starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.log.Info("started workers")
	<-stopCh
	c.log.Info("shutting down workers")
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextItem get the pod from queue and delete the pod
// if it is completed and over ttl seconds
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	defer c.workqueue.Done(obj)

	var key string
	var ok bool
	if key, ok = obj.(string); !ok {
		c.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return true
	}

	pod, err := c.getPod(key)
	if err != nil {
		c.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("pod '%s' in work queue does not exists", key))
		return true
	}

	if pod.Status.Phase != corev1.PodSucceeded {
		c.log.Debugf("this pod is not completed: %s", key)
		return true
	}

	executionTimeSeconds := c.getExecutionTimeSeconds(pod)
	if c.ttl == 0 || c.ttl < executionTimeSeconds {
		c.log.Infof("delete the pod \"%s\", execution time is %d", key, executionTimeSeconds)
		c.deletePod(pod)

		if c.job {
			c.deleteJob(pod)
		}
	}

	c.workqueue.Forget(obj)
	c.log.Infof("processed the item: %s", key)
	return true
}


// enqueuePod enqueue the key if the key(namespace/name) is equal to the pattern
func (c *Controller) enqueuePod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	// match the key and the pattern
	c.log.Debugf("try enqueue the pod to workqueue: %s", key)
	var isMatched bool
	for _, p := range c.patterns {
		if wildcard.MatchSimple(p, key) {
			isMatched = true
			break
		}
	}

	if !isMatched {
		c.log.Debugf("this pod does not match to patterns: %s", key)
		return 
	}
	c.log.Infof("enqueue the key of pod: %s", key)
	c.workqueue.Add(key)
}

func (c *Controller) getPod(key string) (*corev1.Pod, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, fmt.Errorf("invalid resource key: %s", key)
	}

	pod, err := c.lister.Pods(namespace).Get(name)
	return pod, err
}

// method to calculate the hours that passed since the pod's execution end time
func (c *Controller) getExecutionTimeSeconds(pod *corev1.Pod) int {
	now := time.Now()
	for _, pc := range pod.Status.Conditions {
		// Looking for the time when pod's condition "Ready" became "false" (equals end of execution)
		if pc.Type == corev1.PodReady && pc.Status == corev1.ConditionFalse {
			ret := now.Sub(pc.LastTransitionTime.Time).Seconds()
			return int(ret)
		}
	}

	return 0
}

// delete the pod
func (c *Controller) deletePod(pod *corev1.Pod) {
	if c.dryRun {
		c.log.Infof("dry-run: the pod would have been deleted: %s", pod.Name)
		return 
	}

	c.log.Infof("deleting the pod: %s", pod.Name)
	var po metav1.DeleteOptions
	err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &po)
	if err != nil {
		c.log.Infof("failed to delete completed pod: %v",  err)
	}
	return
}

// delete the job if the pod is ownered by the job
func (c *Controller) deleteJob(pod *corev1.Pod) {
	var jobName string
	for _, ow := range pod.OwnerReferences {
		if ow.Kind == "Job" {
			jobName = ow.Name
		}
	}

	if jobName == "" {
		c.log.Infof("this pod did not created by the job: %s", pod.Name)
		return
	}

	if c.dryRun {
		c.log.Infof("dry-run: the job would have been deleted: %s", jobName)
		return 
	}

	c.log.Infof("deleting the job: %s", jobName)
	var jo metav1.DeleteOptions
	err := c.clientset.BatchV1().Jobs(pod.Namespace).Delete(jobName, &jo)
	if err != nil {
		log.Printf("failed to delete job %s: %v", jobName, err)
	}
	return 
}