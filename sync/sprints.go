package sync

import (
	"jiron/db"
	"jiron/jira"
	"log"
)

func Sprints() error {

	client, dbErr := jira.NewSTIPClient()
	if dbErr != nil {
		log.Println(dbErr)
		return dbErr
	}
	jiraSprints, err := client.GetSprintsInBoard(2, []string{"closed", "active", "future"})
	if err != nil {
		log.Println(err)
		return err
	}

	service, err := db.NewSprints()
	if err != nil {
		log.Println(err)
		return err
	}
	err = service.BeginTransaction()
	if err != nil {
		return err
	}
	for _, sprint := range jiraSprints {

		err = service.Upsert(db.Sprint{
			ID:        int16(sprint.ID),
			Name:      sprint.Name,
			State:     sprint.State,
			StartDate: sprint.StartDate,
			EndDate:   sprint.EndDate,
		})
		if err != nil {
			log.Println(err)
			err = service.RollbackTransaction()
			return err
		}
	}
	err = service.CommitTransaction()
	return err
}
