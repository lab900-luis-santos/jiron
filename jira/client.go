package jira

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	j "github.com/andygrunwald/go-jira"
)

const JiraTimeFormat string = "2006-01-02T15:04:05.000Z"

// JiraClient is a wrapper around the go-jira client
type JiraClient struct {
	client *j.Client
}

type Assignee struct {
	Name  string
	Email string
}

type Issue struct {
	Key       string
	Summary   string
	Status    string
	SPs       float64
	CreatedAt time.Time
	Assignee  Assignee
	SyncedOn  time.Time
}

// print all fields in order
func (i Issue) String() string {
	return fmt.Sprintf("Key: %s\nSummary: %s\nSPs: %f\nCreated At: %s\nAssignee: %s\n", i.Key, i.Summary, i.SPs, i.CreatedAt, i.Assignee.String())
}

func (a Assignee) String() string {
	return fmt.Sprintf("Name: %s\nEmail: %s\n", a.Name, a.Email)
}

// Authenticate authenticates with Jira using the API token
func (jc *JiraClient) Authenticate(username, apiToken, baseURL string) error {
	tp := j.BasicAuthTransport{
		Username: username,
		Password: apiToken,
	}

	client, err := j.NewClient(tp.Client(), baseURL)
	if err != nil {
		return err
	}

	jc.client = client
	return nil
}

func (jc *JiraClient) GetCurrentSprintIssues(project string, sprintId int16) ([]Issue, error) {
	var issues []Issue
	syncDate := time.Now()

	// map will convert a jira.Issue to an Issue
	mapIssue := func(i j.Issue) Issue {
		t := time.Time(i.Fields.Created) // convert go-jira.Time to time.Time for manipulation
		assignee := Assignee{}
		if i.Fields.Assignee != nil {
			assignee.Name = i.Fields.Assignee.DisplayName
			assignee.Email = i.Fields.Assignee.EmailAddress
		}
		rawSps := i.Fields.Unknowns["customfield_10016"]
		if rawSps == nil {
			rawSps = 0.0
		}

		var SPs float64 = rawSps.(float64)

		return Issue{
			Key:       i.Key,
			Summary:   i.Fields.Summary,
			Status:    i.Fields.Status.Name,
			SPs:       SPs,
			CreatedAt: t,
			SyncedOn:  syncDate,
			Assignee:  assignee,
		}
	}

	// appendFunc will append jira issues to []jira.Issue
	appendFunc := func(i j.Issue) (err error) {
		issues = append(issues, mapIssue(i))
		return err
	}

	// SearchPages will page through results and pass each issue to appendFunc
	// In this example, we'll search for all the issues in the target project
	var err = jc.client.Issue.SearchPages(fmt.Sprintf(`project=%s AND sprint=%d`, strings.TrimSpace(project), sprintId), nil, appendFunc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d issues found.\n", len(issues))

	return issues, err
}

func (s *JiraClient) GetIssue(key string) (Issue, error) {
	issue, _, err := s.client.Issue.Get(key, nil)
	if err != nil {
		return Issue{}, err
	}
	return Issue{
		Key:       issue.Key,
		Summary:   issue.Fields.Summary,
		Status:    issue.Fields.Status.Name,
		SPs:       issue.Fields.Unknowns["customfield_10016"].(float64),
		CreatedAt: time.Time(issue.Fields.Created),
		Assignee: Assignee{
			Name:  issue.Fields.Assignee.DisplayName,
			Email: issue.Fields.Assignee.EmailAddress,
		},
	}, nil
}

type SprintDto struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Sprint struct {
	ID        int
	Name      string
	State     string
	StartDate time.Time
	EndDate   time.Time
}

func (s *JiraClient) GetSprintsInBoard(boardId int, state []string) ([]Sprint, error) {
	sprintsList, _, err := s.client.Board.GetAllSprintsWithOptions(boardId, &j.GetAllSprintsOptions{State: strings.Join(state[:], ",")})
	if err != nil {
		return nil, err
	}
	var sprints []Sprint

	for _, sprint := range sprintsList.Values {
		sprints = append(sprints, Sprint{ID: sprint.ID, Name: sprint.Name, State: sprint.State})
	}

	return sprints, nil
}

func (s *JiraClient) GetSprint(sprintId int) (*Sprint, error) {
	sprintEndpoint := fmt.Sprintf("rest/agile/1.0/sprint/%d", sprintId)
	req, err := s.client.NewRequestWithContext(context.Background(), "GET", sprintEndpoint, nil)
	sprint := new(SprintDto)
	resp, err := s.client.Do(req, sprint)

	if err != nil {
		jerr := j.NewJiraError(resp, err)
		return nil, jerr
	}
	parsedStart, _ := time.Parse(JiraTimeFormat, sprint.StartDate)
	parsedEnd, _ := time.Parse(JiraTimeFormat, sprint.EndDate)

	return &Sprint{
		ID:        sprint.ID,
		Name:      sprint.Name,
		State:     sprint.State,
		StartDate: parsedStart,
		EndDate:   parsedEnd,
	}, nil
}
