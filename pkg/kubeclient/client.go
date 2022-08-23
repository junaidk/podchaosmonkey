package kubeclient

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Create(configPath string) (*kubernetes.Clientset, error) {

	var clientset *kubernetes.Clientset
	if configPath == "" {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		// creates the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	} else {
		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return nil, err
		}

		// create the clientset
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	}

	return clientset, nil
}
