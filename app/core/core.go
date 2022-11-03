package core

import (
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/procyon-projects/chrono"
	"k8s.io/client-go/kubernetes"
)

type Core struct {
	Projects        Projects
	Services        Services
	ManagedServices ManagedServices
}

func New(storage *storage.Storage, clientset *kubernetes.Clientset, taskScheduler chrono.TaskScheduler) *Core {
	projects := InitProjects(storage, clientset, taskScheduler)
	services := InitServices(projects, storage, clientset)
	projects.(projectsImpl).setServices(services) // TODO refactor
	managedServices := InitManagedServices(projects, storage, clientset)
	return &Core{
		Projects:        projects,
		Services:        services,
		ManagedServices: managedServices,
	}
}
