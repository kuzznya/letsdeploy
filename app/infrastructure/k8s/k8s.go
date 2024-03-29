package k8s

import (
	_ "github.com/abbot/go-http-auth"
	certManagerClientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned/typed/traefikio/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func Setup(cfg *viper.Viper) (*kubernetes.Clientset, *certManagerClientset.Clientset) {
	isInCluster := cfg.GetBool("kubernetes.in-cluster")
	var config *rest.Config
	if isInCluster {
		config = inCluster()
	} else {
		config = outOfCluster(cfg)
	}
	return kubernetes.NewForConfigOrDie(config), certManagerClientset.NewForConfigOrDie(config)
}

func SetupTraefikClient(cfg *viper.Viper) *v1alpha1.TraefikV1alpha1Client {
	isInCluster := cfg.GetBool("kubernetes.in-cluster")
	var config *rest.Config
	if isInCluster {
		config = inCluster()
	} else {
		config = outOfCluster(cfg)
	}
	return v1alpha1.NewForConfigOrDie(config)
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
