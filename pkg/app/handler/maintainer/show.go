package maintainer

import (
	"github.com/expeditioneer/gentoo-soko/pkg/app/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/database"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"github.com/go-pg/pg/v9/orm"
	"net/http"
	"sort"
	"strings"
)

// Show renders a template to show a given maintainer page
func Show(w http.ResponseWriter, r *http.Request) {
	maintainerEmail := r.URL.Path[len("/maintainer/"):]
	maintainerEmail = strings.Split(maintainerEmail, "/")[0]
	if !strings.Contains(maintainerEmail, "@") {
		maintainerEmail = maintainerEmail + "@gentoo.org"
	}

	whereClause := "maintainers @> '[{\"Email\": \"" + maintainerEmail + "\"}]'"
	if maintainerEmail == "maintainer-needed@gentoo.org" {
		whereClause = "maintainers IS null"
	}

	maintainer := models.Maintainer{
		Email: maintainerEmail,
	}
	database.DBCon.Model(&maintainer).WherePK().Relation("Project").Relation("Projects").Select()

	userPreferences := utils.GetUserPreferences(r)
	if userPreferences.Maintainers.IncludeProjectPackages && maintainer.Projects != nil && len(maintainer.Projects) > 0 {
		whereParts := []string{"maintainers @> '[{\"Email\": \"" + maintainerEmail + "\"}]'"}
		for _, proj := range maintainer.Projects {
			if !strings.Contains(strings.Join(userPreferences.Maintainers.ExcludedProjects, ","), proj.Email) {
				whereParts = append(whereParts, "maintainers @> '[{\"Email\": \""+proj.Email+"\"}]'")
			}
		}
		whereClause = strings.Join(whereParts, " OR ")
	}

	var gpackages []*models.Package
	query := database.DBCon.Model(&gpackages).
		Where(whereClause)

	pageName := "packages"
	if strings.HasSuffix(r.URL.Path, "/changelog") {
		pageName = "changelog"
		query = query.
			Relation("Versions").
			Relation("Commits", func(q *orm.Query) (*orm.Query, error) {
				return q.Order("preceding_commits DESC").Limit(50), nil
			})
	} else if strings.HasSuffix(r.URL.Path, "/outdated") {
		pageName = "outdated"
		query = query.
			Relation("Versions").
			Relation("Outdated")
	} else if strings.HasSuffix(r.URL.Path, "/qa-report") {
		pageName = "qa-report"
		query = query.
			Relation("Versions").
			Relation("PkgCheckResults").
			Relation("Versions.PkgCheckResults")
	} else if strings.HasSuffix(r.URL.Path, "/pull-requests") {
		pageName = "pull-requests"
		query = query.
			Relation("Versions").
			Relation("PullRequests")
	} else if strings.HasSuffix(r.URL.Path, "/stabilization") {
		pageName = "stabilization"
		query = query.
			Relation("Versions").
			Relation("PkgCheckResults").
			Relation("Versions.PkgCheckResults").
			Relation("Bugs")
	} else if strings.HasSuffix(r.URL.Path, "/bugs") {
		pageName = "bugs"
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	} else if strings.HasSuffix(r.URL.Path, "/security") {
		pageName = "security"
		query = query.
			Relation("Versions").
			Relation("Versions.Bugs").
			Relation("Bugs")
	} else {
		query = query.
			Relation("Versions").
			Relation("Versions.Masks")
	}

	err := query.Select()

	if err != nil || len(gpackages) == 0 {
		http.NotFound(w, r)
		return
	}

	sort.Slice(gpackages, func(i, j int) bool {
		if gpackages[i].Category != gpackages[j].Category {
			return gpackages[i].Category < gpackages[j].Category
		}
		return gpackages[i].Name < gpackages[j].Name
	})

	renderMaintainerTemplate("show",
		"*",
		GetFuncMap(),
		createMaintainerData(pageName, &maintainer, gpackages),
		w)

}
