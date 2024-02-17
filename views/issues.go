package views

import (
	"html/template"
	"jiron/db"
	"log"
	"net/http"
	"strconv"
)

func SyncIssues(w http.ResponseWriter, r *http.Request) {
	log.Println("Syncing issues")
	// get sprint query param
	sprint := r.URL.Query().Get("sprint")
	// sprint to int16
	intSprint, _ := strconv.Atoi(sprint)
	go db.SyncIssues(int16(intSprint))
}

type IssuesPageData struct {
	PageTitle string
	Issues    []db.Issue
}

func ListDBIssues(w http.ResponseWriter, r *http.Request) {
	service, dbErr := db.NewIssues()
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	defer service.Close()
	issues, _ := service.List()
	tmpl, _ := template.ParseFiles("templates/issues.html")
	data := IssuesPageData{
		PageTitle: "Issues",
		Issues:    issues,
	}
	tmpl.Execute(w, data)
}
