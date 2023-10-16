package projects

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/supagrate/cli/supabase/http"
	"os"
)

type Database struct {
	Host    string `json:"host"`
	Version string `json:"version"`
}

type Project struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organization_id"`
	Name           string   `json:"name"`
	Region         string   `json:"region"`
	Database       Database `json:"database"`
	CreatedAt      string   `json:"created_at"`
}

type Projects = []Project

func ReadProjects(accessToken string) (Projects, error) {
	client := http.NewSupabaseClient(accessToken)

	response, err := client.Get("/projects")

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	var projects Projects
	err = json.Unmarshal(response, &projects)

	if err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

	return projects, nil
}
