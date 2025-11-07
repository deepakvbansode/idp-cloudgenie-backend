package k8s

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeConfig() (*rest.Config, error) {
    // Try in-cluster config first
    config, err := rest.InClusterConfig()
    if err == nil {
        return config, nil
    }
    // Fallback to local kubeconfig
    kubeconfig := os.Getenv("KUBECONFIG")
    if kubeconfig == "" {
        kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
    }
    return clientcmd.BuildConfigFromFlags("", kubeconfig)
}