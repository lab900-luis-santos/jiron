package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"jiron/views"
	"log"
	"net/http"
	"os"
)

func home(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html", "templates/sprint-list.html")
	tmpl.ExecuteTemplate(w, "index", nil)
}

func main() {
	r := mux.NewRouter()
	// static routes
	fs := http.FileServer(http.Dir("assets/"))
	r.Handle("/static/{rest}", http.StripPrefix("/static/", fs))

	// home
	r.HandleFunc("/", home)

	// sprint routes
	r.HandleFunc("/sprint", views.SprintCRUD)
	r.HandleFunc(
		"/sprint/{ulid}",
		views.StoryPointsByStatusAndSyncDate,
	)

	// issues routes
	r.HandleFunc("/issues", views.ListDBIssues)
	r.HandleFunc("/issues/aggregate", views.StoryPointsByStatusAndSyncDate)
	r.HandleFunc("/sync/issues", views.SyncIssues)
	r.HandleFunc("/sync/sprints", views.SyncSprints)

	log.Printf("Starting server at port 8080\n")
	log.Println("Go to http://localhost:8080 to view the application")
	log.Println(fmt.Sprintf("PID: %d", os.Getpid()))
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
