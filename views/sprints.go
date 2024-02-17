package views

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"jiron/db"
	"jiron/jira"
	"jiron/sync"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Sprint struct {
	ULID      string
	ID        int
	Name      string
	State     string
	StartDate string
	EndDate   string
}

type PageData struct {
	Sprints []Sprint
}

func SprintCreateForm(w http.ResponseWriter, r *http.Request) {
	// when request is a get
	if r.Method == "GET" {
		vars := mux.Vars(r)
		id, _ := vars["id"]
		// get status query param
		tmpl, _ := template.ParseFiles("templates/create-sprint.html")
		intId, _ := strconv.Atoi(id)
		client, _ := jira.NewSTIPClient()
		sprint, _ := client.GetSprint(intId)

		err := tmpl.Execute(w, Sprint{
			ID:        sprint.ID,
			Name:      sprint.Name,
			State:     sprint.State,
			StartDate: sprint.StartDate.Format(HTMLTime),
			EndDate:   sprint.EndDate.Format(HTMLTime),
		})
		if err != nil {
			return
		}
	} else if r.Method == "POST" {
		// get sprint values from form post
		r.ParseForm()
		id, _ := strconv.Atoi(r.FormValue("id"))
		name := r.FormValue("name")
		state := r.FormValue("state")
		startDate, _ := time.Parse(HTMLTime, r.FormValue("start"))
		endDate, _ := time.Parse(HTMLTime, r.FormValue("end"))

		sprintService, _ := db.NewSprints()
		sprint := db.Sprint{
			ID:        int16(id),
			Name:      name,
			State:     state,
			StartDate: startDate,
			EndDate:   endDate,
		}
		err := sprintService.Create(sprint)
		if err != nil {
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
}

func SprintCRUD(w http.ResponseWriter, r *http.Request) {

	// when request is a get
	if r.Method == "GET" {
		// get status query param
		status := r.URL.Query().Get("status")

		if status == "active" || status == "future" {
			service, err := db.NewSprints()
			if err != nil {
				log.Println(err)
			}
			defer service.Close()
			dbSprints, _ := service.List([]string{status})
			sprints := make([]Sprint, 0, len(dbSprints))
			for _, s := range dbSprints {
				sprints = append(sprints, Sprint{ULID: s.ULID, ID: int(s.ID), Name: s.Name})
			}

			tmpl, _ := template.ParseFiles(fmt.Sprintf("templates/%s-sprints.html", status))
			data := PageData{
				Sprints: sprints,
			}

			//cache 60s no revalidate
			w.Header().Set("Cache-Control", "public, max-age=60, immutable")

			//return type html
			w.Header().Set("Content-Type", "text/html")

			tmpl.Execute(w, data)
		}
	} else if r.Method == "POST" {
		log.Println("POST NOT IMPLEMENTED")
	}
}

func SyncSprints(w http.ResponseWriter, r *http.Request) {
	go func() {
		err := sync.Sprints()
		if err != nil {
			log.Println(err)
		}
	}()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
