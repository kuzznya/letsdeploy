package core

import (
	"codnect.io/chrono"
	"context"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/app/util/promise"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
)

type Core struct {
	Projects        Projects
	Services        Services
	ManagedServices ManagedServices
	Tokens          Tokens
}

type projectSynchronizable interface {
	syncKubernetes(ctx context.Context, projectId string) error
}

func New(
	storage *storage.Storage,
	rdb *redis.Client,
	clientset *kubernetes.Clientset,
	taskScheduler chrono.TaskScheduler,
	cfg *viper.Viper,
) *Core {
	corePromise := promise.New[Core]()
	projects := InitProjects(storage, clientset, cfg, corePromise)
	services := InitServices(projects, storage, clientset, cfg)
	managedServices := InitManagedServices(projects, storage, clientset)
	tokens := InitTokens(rdb)
	core := &Core{
		Projects:        projects,
		Services:        services,
		ManagedServices: managedServices,
		Tokens:          tokens,
	}
	corePromise.Resolve(*core)
	InitSync(core, taskScheduler)
	return core
}
