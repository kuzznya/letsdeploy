package core

import (
	"codnect.io/chrono"
	"context"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/app/util/promise"
	"k8s.io/client-go/kubernetes"
)

type Core struct {
	Projects        Projects
	Services        Services
	ManagedServices ManagedServices
}

type projectSynchronizable interface {
	syncKubernetes(ctx context.Context, projectId string) error
}

func New(storage *storage.Storage, clientset *kubernetes.Clientset, taskScheduler chrono.TaskScheduler) *Core {
	corePromise := promise.New[Core]()
	projects := InitProjects(storage, clientset, corePromise)
	services := InitServices(projects, storage, clientset)
	managedServices := InitManagedServices(projects, storage, clientset)
	core := &Core{
		Projects:        projects,
		Services:        services,
		ManagedServices: managedServices,
	}
	corePromise.Resolve(*core)
	InitSync(core, taskScheduler)
	return core
}
