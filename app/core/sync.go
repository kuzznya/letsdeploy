package core

import (
	"codnect.io/chrono"
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

func InitSync(core *Core, scheduler chrono.TaskScheduler) {
	_, err := scheduler.ScheduleWithFixedDelay(sync(core), 1*time.Minute)
	if err != nil {
		log.WithError(err).Panicln("Unable to schedule k8s synchronization")
	}
}

func sync(core *Core) chrono.Task {
	return func(ctx context.Context) {
		log.Infoln("Kubernetes sync started")

		limit := 1000
		offset := 0
		checkedProjects := make(map[string]bool)
		for {
			projects, err := core.Projects.FindAll(limit, offset)
			if err != nil {
				log.WithError(err).Errorln("Failed to retrieve projects")
				return
			}
			for _, project := range projects {
				checkedProjects[project.Id] = true

				if err := core.Projects.syncKubernetes(ctx, project.Id); err != nil {
					log.WithError(err).Errorf("Project %s sync failed, skipping", project.Id)
					continue
				}
				if err := core.Registries.syncKubernetes(ctx, project.Id); err != nil {
					log.WithError(err).Errorf("Project %s registries sync failed", project.Id)
				}
				if err := core.Services.syncKubernetes(ctx, project.Id); err != nil {
					log.WithError(err).Errorf("Project %s services sync failed", project.Id)
				}
				if err := core.ManagedServices.syncKubernetes(ctx, project.Id); err != nil {
					log.WithError(err).Errorf("Project %s managed services sync failed", project.Id)
				}
			}
			if len(projects) < limit {
				break
			}
		}

		core.Projects.(*projectsImpl).removeExcessNamespaces(ctx, checkedProjects) // TODO: 09.11.22 refactor

		log.Infoln("Kubernetes sync finished")
	}
}
