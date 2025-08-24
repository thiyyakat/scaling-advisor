package clientutil

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

// BuildClients currently constructs static and dynamic client-go clients given a kubeconfig file.
func BuildClients(kubeConfigPath string) (client kubernetes.Interface, dynClient dynamic.Interface, err error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return
	}
	clientConfig.ContentType = "application/json"
	client, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return
	}
	dynClient, err = dynamic.NewForConfig(clientConfig)
	if err != nil {
		return
	}
	return
}

func BuildInformerFactories(client kubernetes.Interface, dyncClient dynamic.Interface, resyncPeriod time.Duration) (informerFactory informers.SharedInformerFactory, dynInformerFactory dynamicinformer.DynamicSharedInformerFactory) {
	informerFactory = informers.NewSharedInformerFactory(client, resyncPeriod)
	dynInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(dyncClient, resyncPeriod)
	return
}
