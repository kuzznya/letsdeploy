package core

import (
	"cmp"
	"context"
	"fmt"
	"github.com/kuzznya/letsdeploy/app/apperrors"
	"github.com/kuzznya/letsdeploy/app/middleware"
	"github.com/kuzznya/letsdeploy/app/storage"
	"github.com/kuzznya/letsdeploy/internal/openapi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/client-go/kubernetes"
	"slices"
)

type MongoDbMgmt interface {
	GetMongoDbUsers(ctx context.Context, serviceId int, auth middleware.Authentication) ([]openapi.MongoDbUser, error)
	GetMongoDbUser(ctx context.Context, serviceId int, mongoDbUsername string, auth middleware.Authentication) (openapi.MongoDbUser, error)
	CreateMongoDbUser(ctx context.Context, serviceId int, mongoDbUser openapi.MongoDbUser, auth middleware.Authentication) (openapi.MongoDbUser, error)
	UpdateMongoDbUser(ctx context.Context, serviceId int, mongoDbUser openapi.MongoDbUser, authentication middleware.Authentication) (openapi.MongoDbUser, error)
	DeleteMongoDbUser(ctx context.Context, serviceId int, mongoDbUsername string, auth middleware.Authentication) error
}

type mongoDbMgmtImpl struct {
	managedServices ManagedServices
	storage         *storage.Storage
	clientset       *kubernetes.Clientset
}

var _ MongoDbMgmt = (*mongoDbMgmtImpl)(nil)

type userInfo struct {
	User  string         `bson:"user"`
	Roles []userInfoRole `bson:"roles,omitempty"`
}

type userInfoRole struct {
	Role string `bson:"role"`
	Db   string `bson:"db"`
}

type userInfoResp struct {
	Users []userInfo `bson:"users"`
}

func InitMongoDbMgmt(
	managedServices ManagedServices,
	storage *storage.Storage,
	clientset *kubernetes.Clientset,
) MongoDbMgmt {
	return &mongoDbMgmtImpl{managedServices: managedServices, storage: storage, clientset: clientset}
}

func (m mongoDbMgmtImpl) GetMongoDbUsers(ctx context.Context, serviceId int, auth middleware.Authentication) ([]openapi.MongoDbUser, error) {
	service, err := m.getMongoDbService(ctx, serviceId, auth)
	if err != nil {
		return nil, err
	}

	client, err := m.getMongoDbClient(ctx, *service)
	if err != nil {
		return nil, err
	}

	cmd := bson.D{{"usersInfo", 1}} // MongoDB does not allow to retrieve privileges for all users
	resp := userInfoResp{}
	err = client.Database("admin").RunCommand(ctx, cmd).Decode(&resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get MongoDB users")
	}

	users := make([]openapi.MongoDbUser, 0)
	for _, user := range resp.Users {
		if user.User == managedServices[openapi.Mongo].username {
			continue
		}

		u, err := m.getMongoDbUserInfo(ctx, client, user.User)
		if err != nil && !apperrors.IsNotFound(err) {
			return nil, err
		}
		users = append(users, u)
	}
	slices.SortFunc(users, func(a, b openapi.MongoDbUser) int {
		return cmp.Compare(a.Username, b.Username)
	})
	log.Debugf("Retrieved MongoDB users of service %s in project %s", service.Name, service.Project)
	return users, nil
}

func (m mongoDbMgmtImpl) GetMongoDbUser(ctx context.Context, serviceId int, mongoDbUsername string, auth middleware.Authentication) (openapi.MongoDbUser, error) {
	service, err := m.getMongoDbService(ctx, serviceId, auth)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	client, err := m.getMongoDbClient(ctx, *service)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	u, err := m.getMongoDbUserInfo(ctx, client, mongoDbUsername)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}
	if u.Username == managedServices[openapi.Mongo].username {
		return openapi.MongoDbUser{}, apperrors.NotFound("User " + u.Username + " not found")
	}

	log.Debugf("Retrieved MongoDB user of service %s in project %s", service.Name, service.Project)

	return u, nil
}

func (m mongoDbMgmtImpl) CreateMongoDbUser(ctx context.Context, serviceId int, mongoDbUser openapi.MongoDbUser, auth middleware.Authentication) (openapi.MongoDbUser, error) {
	service, err := m.getMongoDbService(ctx, serviceId, auth)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	client, err := m.getMongoDbClient(ctx, *service)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	if mongoDbUser.Username == managedServices[openapi.Mongo].username {
		return openapi.MongoDbUser{}, apperrors.Forbidden("Cannot create user with username 'root'")
	}

	_, err = m.getMongoDbUserInfo(ctx, client, mongoDbUser.Username)
	if err == nil || !apperrors.IsNotFound(err) {
		return openapi.MongoDbUser{}, apperrors.BadRequest("User " + mongoDbUser.Username + " already exists")
	}

	if mongoDbUser.PasswordSecret == nil {
		return openapi.MongoDbUser{}, apperrors.BadRequest("passwordSecret should be provided")
	}

	secret, err := m.storage.SecretRepository().FindByProjectIdAndName(service.Project, *mongoDbUser.PasswordSecret)
	if err != nil {
		return openapi.MongoDbUser{}, errors.Wrap(err, "failed to get corresponding secret")
	}

	roles := mapItems[openapi.MongoDbRole, userInfoRole](mongoDbUser.Roles, func(r openapi.MongoDbRole) userInfoRole {
		return userInfoRole{Role: string(r.Role), Db: r.Db}
	})

	cmd := bson.D{{"createUser", mongoDbUser.Username}, {"pwd", secret.Value}, {"roles", roles}}
	err = client.Database("admin").RunCommand(ctx, cmd).Err()
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return openapi.MongoDbUser{}, apperrors.BadRequest("User " + mongoDbUser.Username + " already exists")
		}
		return openapi.MongoDbUser{}, errors.Wrap(err, "failed to create user")
	}
	log.Infof("Created MongoDB user of service %s in project %s", service.Name, service.Project)
	return mongoDbUser, nil
}

func (m mongoDbMgmtImpl) UpdateMongoDbUser(ctx context.Context, serviceId int, mongoDbUser openapi.MongoDbUser, auth middleware.Authentication) (openapi.MongoDbUser, error) {
	service, err := m.getMongoDbService(ctx, serviceId, auth)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	client, err := m.getMongoDbClient(ctx, *service)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	_, err = m.getMongoDbUserInfo(ctx, client, mongoDbUser.Username)
	if err != nil {
		return openapi.MongoDbUser{}, err
	}

	roles := mapItems[openapi.MongoDbRole, userInfoRole](mongoDbUser.Roles, func(r openapi.MongoDbRole) userInfoRole {
		return userInfoRole{Role: string(r.Role), Db: r.Db}
	})

	var cmd bson.D
	if mongoDbUser.PasswordSecret != nil {
		secret, err := m.storage.SecretRepository().FindByProjectIdAndName(service.Project, *mongoDbUser.PasswordSecret)
		if err != nil {
			return openapi.MongoDbUser{}, errors.Wrap(err, "failed to get corresponding secret")
		}
		cmd = bson.D{{"updateUser", mongoDbUser.Username}, {"pwd", secret.Value}, {"roles", roles}}
	} else {
		cmd = bson.D{{"updateUser", mongoDbUser.Username}, {"roles", roles}}
	}

	err = client.Database("admin").RunCommand(ctx, cmd).Err()
	if err != nil {
		return openapi.MongoDbUser{}, errors.Wrap(err, "failed to update user")
	}
	log.Infof("Updated MongoDB user of service %s in project %s", service.Name, service.Project)
	return mongoDbUser, nil
}

func (m mongoDbMgmtImpl) DeleteMongoDbUser(ctx context.Context, serviceId int, mongoDbUsername string, auth middleware.Authentication) error {
	service, err := m.getMongoDbService(ctx, serviceId, auth)
	if err != nil {
		return err
	}

	client, err := m.getMongoDbClient(ctx, *service)
	if err != nil {
		return err
	}

	_, err = m.getMongoDbUserInfo(ctx, client, mongoDbUsername)
	if err != nil && apperrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if mongoDbUsername == managedServices[openapi.Mongo].username {
		return apperrors.Forbidden("Cannot delete user with username 'root'")
	}

	cmd := bson.D{{"dropUser", mongoDbUsername}}
	err = client.Database("admin").RunCommand(ctx, cmd).Err()
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	log.Infof("Deleted MongoDB user of service %s in project %s", service.Name, service.Project)
	return nil
}

func (m mongoDbMgmtImpl) getMongoDbService(ctx context.Context, serviceId int, auth middleware.Authentication) (*openapi.ManagedService, error) {
	service, err := m.managedServices.GetManagedService(serviceId, auth)
	if err != nil {
		return nil, err
	}
	status, err := m.managedServices.GetManagedServiceStatus(ctx, serviceId, auth)
	if err != nil {
		return nil, err
	}
	if service.Type != openapi.Mongo {
		return nil, apperrors.BadRequest("Managed service is not MongoDB")
	}
	if status.Status != openapi.Available {
		return nil, apperrors.BadRequest("MongoDB is not available")
	}
	return service, nil
}

func (m mongoDbMgmtImpl) getMongoDbClient(ctx context.Context, service openapi.ManagedService) (*mongo.Client, error) {
	mongoHost := fmt.Sprintf("%s.%s.svc.cluster.local:%d",
		service.Name, service.Project, managedServices[openapi.Mongo].podPort)
	secret, err := m.storage.SecretRepository().FindByProjectIdAndName(service.Project, getManagedServiceSecretName(service.Name))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get root password for MongoDB")
	}
	credential := options.Credential{Username: managedServices[openapi.Mongo].username, Password: secret.Value}
	clientOptions := options.Client().SetHosts([]string{mongoHost}).SetAuth(credential)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to MongoDB")
	}
	return client, nil
}

func (m mongoDbMgmtImpl) getMongoDbUserInfo(ctx context.Context, client *mongo.Client, username string) (openapi.MongoDbUser, error) {
	cmd := bson.D{{"usersInfo", username}, {"showPrivileges", true}}
	resp := userInfoResp{}
	err := client.Database("admin").RunCommand(ctx, cmd).Decode(&resp)
	u := resp.Users
	if err != nil {
		return openapi.MongoDbUser{}, errors.Wrap(err, "failed to get MongoDB user")
	}
	if len(u) == 0 {
		return openapi.MongoDbUser{}, apperrors.NotFound("User " + username + " not found")
	}
	user := u[0]

	roles := make([]openapi.MongoDbRole, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = openapi.MongoDbRole{
			Db:   role.Db,
			Role: openapi.MongoDbRoleRole(role.Role),
		}
	}
	return openapi.MongoDbUser{Username: user.User, Roles: roles}, nil
}
