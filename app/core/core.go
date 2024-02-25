package core

import (
	"codnect.io/chrono"
	"context"
	certManagerClientset "github.com/cert-manager/cert-manager/pkg/client/clientset/versioned"
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
	MongoDbMgmt     MongoDbMgmt
	Tokens          Tokens
	ApiKeys         ApiKeys
}

type projectSynchronizable interface {
	syncKubernetes(ctx context.Context, projectId string) error
}

func New(
	storage *storage.Storage,
	rdb *redis.Client,
	clientset *kubernetes.Clientset,
	cmClient *certManagerClientset.Clientset,
	taskScheduler chrono.TaskScheduler,
	cfg *viper.Viper,
) *Core {
	corePromise := promise.New[Core]()
	projects := InitProjects(storage, clientset, cmClient, cfg, corePromise)
	services := InitServices(projects, storage, clientset, cfg)
	managedServices := InitManagedServices(projects, storage, clientset)
	mongoDbMgmt := InitMongoDbMgmt(managedServices, storage, clientset)
	tokens := InitTokens(rdb)
	apiKeys := InitApiKeys(storage)

	core := &Core{
		Projects:        projects,
		Services:        services,
		ManagedServices: managedServices,
		MongoDbMgmt:     mongoDbMgmt,
		Tokens:          tokens,
		ApiKeys:         apiKeys,
	}
	corePromise.Resolve(*core)
	InitSync(core, taskScheduler)
	return core
}
