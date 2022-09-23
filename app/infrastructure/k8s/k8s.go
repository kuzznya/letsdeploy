package k8s

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func Setup(cfg *viper.Viper) *kubernetes.Clientset {
	isInCluster := cfg.GetBool("kubernetes.in-cluster")
	var config *rest.Config
	if isInCluster {
		config = inCluster()
	} else {
		config = outOfCluster(cfg)
	}
	return kubernetes.NewForConfigOrDie(config)
}

func inCluster() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithError(err).Panicln("failed to create in-cluster config")
	}
	return config
}

func outOfCluster(cfg *viper.Viper) *rest.Config {
	masterUrl := cfg.GetString("kubernetes.master-url")
	var kubeconfig string
	if kubeconfigEnv, found := os.LookupEnv("KUBECONFIG"); found {
		kubeconfig = kubeconfigEnv
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else if masterUrl == "" {
		log.Panicln("Please define either kubernetes.master-url, or KUBECONFIG env var, " +
			"or HOME env var to use $HOME/.kube/config")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfig)
	if err != nil {
		log.WithError(err).Panicln("failed to create out-of-cluster config", config)
	}
	return config
}
