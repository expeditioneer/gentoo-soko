package maintainers

import (
	"soko/pkg/database"
	"soko/pkg/models"
	"strings"
)

func FullImport() {

	database.Connect()
	defer database.DBCon.Close()

	var allMaintainerInformation []*models.Maintainer
	database.DBCon.Model((*models.Package)(nil)).ColumnExpr("jsonb_array_elements(maintainers)->>'Name' as name, jsonb_array_elements(maintainers) ->> 'Email' as email, jsonb_array_elements(maintainers) ->> 'Type' as type").Select(&allMaintainerInformation)

	maintainers := map[string]*models.Maintainer{}

	for _, rawMaintainer := range allMaintainerInformation {
		_, ok := maintainers[rawMaintainer.Email]
		if !ok {
			maintainers[rawMaintainer.Email] = rawMaintainer
		} else {
			if maintainers[rawMaintainer.Email].Name == "" {
				maintainers[rawMaintainer.Email].Name = rawMaintainer.Name
			}
		}
	}

	var gpackages []*models.Package
	database.DBCon.Model(&gpackages).
		Relation("Outdated").
		Relation("PullRequests").
		Relation("Bugs").
		Select()

	// TODO in future we want an incremental update here
	// but for now we delete everything and insert it again
	// this is currently acceptable as it takes less than 2 seconds
	deleteAllMaintainers()

	for _, maintainer := range maintainers {
		outdated := 0
		pullRequests := 0
		securityBugs := 0
		nonSecurityBugs := 0

		for _, gpackage := range gpackages {
			found := false
			for _, packageMaintainer := range gpackage.Maintainers {
				if packageMaintainer.Email == maintainer.Email {
					found = true
				}
			}

			if found {
				outdated = outdated + len(gpackage.Outdated)
				pullRequests = pullRequests + len(gpackage.PullRequests)
				for _, bug := range gpackage.Bugs {
					if bug.Component == "Vulnerabilities" {
						securityBugs++
					} else {
						nonSecurityBugs++
					}
				}

			}
		}

		maintainer.PackagesInformation = models.MaintainerPackagesInformation{
			Outdated:     outdated,
			PullRequests: pullRequests,
			Bugs:         nonSecurityBugs,
			SecurityBugs: securityBugs,
		}

		if maintainer.Name == "" {
			maintainer.Name = strings.Title(strings.Split(maintainer.Email, "@")[0])
		}

		if maintainer.Type == "project" && strings.HasPrefix(maintainer.Name, "Gentoo ") {
			maintainer.Name = strings.Replace(maintainer.Name, "Gentoo ", "", 1)
		}

		if maintainer.Type == "person" {
			if strings.HasSuffix(maintainer.Email, "@gentoo.org") {
				maintainer.Type = "gentoo-developer"
			} else {
				maintainer.Type = "proxied-maintainer"
			}
		}

		database.DBCon.Model(maintainer).WherePK().OnConflict("(email) DO UPDATE").Insert()
	}

}

// deleteAllMaintainers deletes all entries in the maintainers table
func deleteAllMaintainers() {
	var maintainers []*models.Maintainer
	database.DBCon.Model(&maintainers).Select()
	for _, maintainer := range maintainers {
		database.DBCon.Model(maintainer).WherePK().Delete()
	}
}
