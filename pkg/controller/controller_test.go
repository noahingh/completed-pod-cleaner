package controller

import (
	"time"
	"testing"

	kubeFake "k8s.io/client-go/kubernetes/fake"
	kubeInformers "k8s.io/client-go/informers"
	// "k8s.io/client-go/tools/cache"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	clientset 	 *kubeFake.Clientset
	// objects is the list of pod
	objects []runtime.Object
}


func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	return f
}

func newPod(name string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name: "hello-world",
					Image: "hello-world",
				},
			},
		},
	}
}

func (f *fixture)newController(patterns []string) (*Controller, kubeInformers.SharedInformerFactory) {
	f.clientset = kubeFake.NewSimpleClientset(f.objects...)
	factory := kubeInformers.NewSharedInformerFactory(f.clientset, noResyncPeriodFunc())

	f.t.Log("Create a new controller")
	c := NewController(
		f.clientset, 
		factory.Core().V1().Pods(),
		patterns, 
		0, 
		true,
		true,
	)
	c.synced = alwaysReady

	f.t.Log("Add objects into the indexer")
	for _, o := range f.objects {
		p := o.(*corev1.Pod)
		factory.Core().V1().Pods().Informer().GetIndexer().Add(p)
	}

	return c, factory
}


func TestEnqueue(t *testing.T) {
	f := newFixture(t)
	c, _ := f.newController([]string{
		"default/hell*",
	})

	f.t.Log("Create the pod")
	p := newPod("hello")

	f.t.Log("Enqueue the pod")
	c.enqueuePod(p)

	f.t.Logf("The size of the queue is %d", c.workqueue.Len())
	if c.workqueue.Len() != 1 {
		f.t.Errorf("Enqueue the pod has failed")
	}
}