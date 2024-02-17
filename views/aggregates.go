package views

import (
	"github.com/gorilla/mux"
	"html/template"
	"jiron/db"
	"log"
	"net/http"
)

type Dataset struct {
	Label       string    `json:"label"`
	Data        []float64 `json:"data"`
	BorderWidth int8      `json:"borderWidth"`
}

type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"data"`
}

func StoryPointsByStatusAndSyncDate(w http.ResponseWriter, r *http.Request) {
	service, dbErr := db.NewIssues()
	if dbErr != nil {
		http.Error(w, dbErr.Error(), http.StatusInternalServerError)
		return
	}
	defer service.Close()

	//get ulid from path
	vars := mux.Vars(r)
	ulid, _ := vars["ulid"]

	aggregates, err := service.StoryPointsByStatusAndSyncDate(ulid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	dataset := make(map[string]Dataset)

	for _, aggregate := range aggregates {
		if d, found := dataset[aggregate.Status]; !found {
			d = Dataset{Label: aggregate.Status, Data: []float64{}, BorderWidth: 1}
			d.Data = append(dataset[aggregate.Status].Data, aggregate.TotalStoryPoints)
			dataset[aggregate.Status] = d
		} else {
			d.Data = append(dataset[aggregate.Status].Data, aggregate.TotalStoryPoints)
			dataset[aggregate.Status] = d
		}
	}

	data := make([]Dataset, 0, len(dataset))
	for _, v := range dataset {
		data = append(data, v)
	}

	tmpl, _ := template.ParseFiles("templates/chart.html")

	labels := make([]string, 0, len(aggregates))
	// find all unique sync dates and add them to labels
	for _, aggregate := range aggregates {
		found := false
		for _, label := range labels {
			if label == aggregate.SyncedOn.Format("15:04:05 02 Jan 2006") {
				found = true
				break
			}
		}
		if !found {
			labels = append(labels, aggregate.SyncedOn.Format("15:04:05 02 Jan 2006"))
		}
	}

	pageData := ChartData{
		Labels:   labels,
		Datasets: data,
	}
	tmpl.Execute(w, pageData)
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(aggregates)
}
