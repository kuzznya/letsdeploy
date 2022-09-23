package core

import (
	"github.com/kuzznya/letsdeploy/app/storage"
	"k8s.io/client-go/kubernetes"
)

type Core struct {
	Projects Projects
}

func New(storage *storage.Storage, clientset *kubernetes.Clientset) *Core {
	projects := Projects{storage: storage, clientset: clientset}
	return &Core{Projects: projects}
}
