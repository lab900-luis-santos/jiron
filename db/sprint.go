package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	ulid "github.com/oklog/ulid/v2"
)

type Sprint struct {
	ULID      string
	ID        int16
	Name      string
	State     string
	StartDate time.Time
	EndDate   time.Time
}

type SprintService struct {
	db *sql.DB
}

const createSprintsTable string = `
CREATE TABLE IF NOT EXISTS sprint (
	ulid TEXT PRIMARY KEY,
	id INTEGER UNIQUE,
	name TEXT,
	state TEXT,
	start_date TEXT,
	end_date TEXT
)
`

func NewSprints() (*SprintService, error) {
	db, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createSprintsTable)
	if err != nil {
		return nil, err
	}

	return &SprintService{db}, nil
}

func (s *SprintService) Close() {
	s.db.Close()
}

func (s *SprintService) Create(sprint Sprint) error {
	_, err := s.db.Exec("INSERT INTO sprint (ulid, id, name, state, start_date, end_date) VALUES (?, ?, ?, ?, ?, ?)",
		ulid.Make().String(), sprint.ID, sprint.Name, sprint.State, sprint.StartDate.Format(Time), sprint.EndDate.Format(Time))
	return err
}

func (s *SprintService) List(state []string) ([]Sprint, error) {
	// filter by state if provided
	filter := ""
	if len(state) > 0 {
		filter = " WHERE state IN ("
		for i, s := range state {
			if i == 0 {
				filter += "'" + s + "'"
			} else {
				filter += ", '" + s + "'"
			}
		}
		filter += ")"
	}
	rows, err := s.db.Query("SELECT ulid, id, name, state, start_date, end_date FROM sprint" + filter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		var sprint Sprint
		var startDate string
		var endDate string
		err = rows.Scan(&sprint.ULID, &sprint.ID, &sprint.Name, &sprint.State, &startDate, &endDate)
		if err != nil {
			return nil, err
		}
		parsedStartDate, err := time.Parse(Time, startDate)
		if err != nil {
			return nil, err
		}
		parsedEndDate, err := time.Parse(Time, endDate)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, Sprint{sprint.ULID, sprint.ID, sprint.Name, sprint.State, parsedStartDate, parsedEndDate})
	}

	return sprints, nil
}

func (s *SprintService) BeginTransaction() error {
	_, err := s.db.Exec("BEGIN TRANSACTION")
	return err
}

func (s *SprintService) CommitTransaction() error {
	_, err := s.db.Exec("COMMIT TRANSACTION")
	return err
}

func (s *SprintService) RollbackTransaction() error {
	_, err := s.db.Exec("ROLLBACK TRANSACTION")
	return err
}

func (s *SprintService) Upsert(sprint Sprint) error {
	// check if sprint exists by id
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sprint WHERE id = ?", sprint.ID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		// insert
		_, err = s.db.Exec("INSERT INTO sprint (ulid, id, name, state, start_date, end_date) VALUES (?, ?, ?, ?, ?, ?)",
			ulid.Make().String(), sprint.ID, sprint.Name, sprint.State, sprint.StartDate.Format(Time), sprint.EndDate.Format(Time))
		if err != nil {
			return err
		}
	} else {
		// update
		_, err = s.db.Exec("UPDATE sprint SET name = ?, state = ?, start_date = ?, end_date = ? WHERE id = ?",
			sprint.Name, sprint.State, sprint.StartDate.Format(Time), sprint.EndDate.Format(Time), sprint.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SprintService) Get(id int16) (*Sprint, error) {
	var sprint Sprint
	var startDate string
	var endDate string
	err := s.db.QueryRow("SELECT ulid, id, name, state, start_date, end_date FROM sprint WHERE id = ?", id).Scan(&sprint.ULID, &sprint.ID, &sprint.Name, &sprint.State, &startDate, &endDate)
	if err != nil {
		return nil, err
	}
	parsedStartDate, err := time.Parse(Time, startDate)
	if err != nil {
		return nil, err
	}
	parsedEndDate, err := time.Parse(Time, endDate)
	if err != nil {
		return nil, err
	}
	sprint.StartDate = parsedStartDate
	sprint.EndDate = parsedEndDate
	return &sprint, nil
}
