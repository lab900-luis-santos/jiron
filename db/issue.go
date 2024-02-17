package db

import (
	"database/sql"
	"fmt"
	"jiron/jira"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	ulid "github.com/oklog/ulid/v2"
)

type Assignee struct {
	Name  string
	Email string
}

// print all fields in order
func (a Assignee) String() string {
	return fmt.Sprintf("Name: %s\nEmail: %s\n", a.Name, a.Email)
}

type Issue struct {
	Key         string
	Summary     string
	StoryPoints float64
	Status      string
	CreatedAt   time.Time
	Assignee    Assignee
	SyncedOn    time.Time
	SprintID    string
}

// print all fields in order
func (i Issue) String() string {
	return fmt.Sprintf("Key: %s\nSummary: %s\nSPs: %f\nCreated At: %s\nAssignee: %s\n", i.Key, i.Summary, i.StoryPoints, i.CreatedAt, i.Assignee.String())
}

type IssueService struct {
	db *sql.DB
}

const createIssueTable string = `
CREATE TABLE IF NOT EXISTS issues (
	id TEXT PRIMARY KEY,
	key TEXT, 
	summary TEXT,
	status TEXT, 
	story_points REAL, 
	created_at TEXT, 
	assignee_name TEXT, 
	assignee_email TEXT,
	sprint_id TEXT,
	synced_on DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY(sprint_id) REFERENCES sprint(ulid)
)
`

func NewIssues() (*IssueService, error) {
	db, err := sql.Open("sqlite3", DBName)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createIssueTable)
	if err != nil {
		return nil, err
	}

	return &IssueService{db: db}, nil
}

func (is *IssueService) Close() {
	is.db.Close()
}

func (is *IssueService) Save(i Issue) error {
	_, err := is.db.Exec("INSERT INTO issues (id, key, summary, status, story_points, created_at, assignee_name, assignee_email, synced_on, sprint_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", ulid.Make().String(), i.Key, i.Summary, i.Status, i.StoryPoints, i.CreatedAt.Format(Time), i.Assignee.Name, i.Assignee.Email, i.SyncedOn.Format(Time), i.SprintID)
	if err != nil {
		log.Print(err)
	}
	return err
}

func (is *IssueService) List() ([]Issue, error) {
	rows, err := is.db.Query("SELECT key, summary, story_points, created_at, assignee_name, assignee_email, synced_on FROM issues")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var key string
		var summary string
		var storyPoints float64
		var createdAt string // Change the type to string
		var assigneeName string
		var assigneeEmail string
		var syncedOn string
		err := rows.Scan(&key, &summary, &storyPoints, &createdAt, &assigneeName, &assigneeEmail, &syncedOn)
		if err != nil {
			return nil, err
		}
		createdAtTime, err := time.Parse(Time, createdAt)
		if err != nil {
			log.Print(err)
		}
		syncedOnTime, err := time.Parse(Time, syncedOn)
		if err != nil {
			log.Print(err)
		}
		issues = append(issues, Issue{Key: key, Summary: summary, StoryPoints: storyPoints, CreatedAt: createdAtTime, SyncedOn: syncedOnTime, Assignee: Assignee{Name: assigneeName, Email: assigneeEmail}})
	}

	return issues, nil
}

type StoryPoint struct {
	Status           string
	SyncedOn         time.Time
	TotalStoryPoints float64
}

func (is *IssueService) StoryPointsByStatusAndSyncDate(sprint string) ([]StoryPoint, error) {
	rows, err := is.db.Query(`
	SELECT status, synced_on, SUM(story_points) AS total_story_points
	FROM issues
	WHERE sprint_id = ?
	GROUP BY status, synced_on
	ORDER BY synced_on ASC`, sprint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var storyPoints []StoryPoint
	for rows.Next() {
		var status string
		var syncedOn string
		var totalStoryPoints float64
		err := rows.Scan(&status, &syncedOn, &totalStoryPoints)
		if err != nil {
			return nil, err
		}
		syncedOnTime, err := time.Parse(Time, syncedOn)
		if err != nil {
			log.Print(err)
		}
		storyPoints = append(storyPoints, StoryPoint{Status: status, SyncedOn: syncedOnTime, TotalStoryPoints: totalStoryPoints})
	}
	return storyPoints, nil
}

func SyncIssues(sprintId int16) {
	client, err := jira.NewSTIPClient()
	issues, err := client.GetCurrentSprintIssues("STIP", sprintId)
	sprintService, err := NewSprints()
	if err != nil {
		log.Print(err)
	} else {
		defer sprintService.Close()
		sprint, err := sprintService.Get(sprintId)
		if err != nil {
			log.Print(err)
		}
		log.Printf("Total Issues: %d\n", len(issues)) // Fix: Changed format specifier from %f to %d
		service, dbErr := NewIssues()
		if dbErr != nil {
			log.Fatal(dbErr)
		}
		defer service.Close()
		for _, i := range issues {
			issue := Issue{
				Key:         i.Key,
				Summary:     i.Summary,
				Status:      i.Status,
				StoryPoints: i.SPs,
				CreatedAt:   i.CreatedAt,
				SyncedOn:    i.SyncedOn,
				SprintID:    sprint.ULID,
				Assignee: Assignee{
					Name:  i.Assignee.Name,
					Email: i.Assignee.Email,
				},
			}
			service.Save(issue)
		}
		log.Print("Issues saved to database\n")
	}
}
