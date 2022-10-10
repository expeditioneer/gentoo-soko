package projects

import (
	"encoding/xml"
	"github.com/expeditioneer/gentoo-soko/internal/config"
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/logger"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"io"
	"log"
	"net/http"
	"time"
)

// UpdatePkgCheckResults will update the database table that contains all pkgcheck results
func UpdateProjects() {

	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(io.Discard)
	}

	// get the internal check results from qa-reports.gentoo.org
	projectList, err := parseProjectList()
	if err != nil {
		logger.Error.Println("Error while parsing project list. Aborting...")
	}

	// clean up the database
	deleteAllProjects()

	// insert new project list
	for _, project := range projectList.Projects {
		database.DBCon.Insert(&project)
		for _, member := range project.Members {
			database.DBCon.Insert(&models.MaintainerToProject{
				Id:              member.Email + "-" + project.Email,
				MaintainerEmail: member.Email,
				ProjectEmail:    project.Email,
			})
		}
	}

	updateStatus()
}

// parseQAReport gets the xml from qa-reports.gentoo.org and parses it
func parseProjectList() (models.ProjectList, error) {
	resp, err := http.Get("https://api.gentoo.org/metastructure/projects.xml")
	if err != nil {
		return models.ProjectList{}, err
	}
	defer resp.Body.Close()
	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.ProjectList{}, err
	}
	var projectList models.ProjectList
	xml.Unmarshal(xmlData, &projectList)
	return projectList, err
}

// deleteAllOutdated deletes all entries in the outdated table
func deleteAllProjects() {
	var allProjects []*models.Project
	database.DBCon.Model(&allProjects).Select()
	for _, project := range allProjects {
		database.DBCon.Model(project).WherePK().Delete()
	}

	var allMaintainerToProjects []*models.MaintainerToProject
	database.DBCon.Model(&allMaintainerToProjects).Select()
	for _, maintainerToProject := range allMaintainerToProjects {
		database.DBCon.Model(maintainerToProject).WherePK().Delete()
	}
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "projects",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
