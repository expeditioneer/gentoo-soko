// Used to search for packages

package packages

import (
	"github.com/expeditioneer/gentoo-soko/internal/app/handler/feeds"
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"github.com/go-pg/pg"
	"net/http"
	"strings"
)

// Search renders a template containing a list of search results
// for a given query of packages
func Search(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)
	var packages []models.Package
	var err error

	if strings.Contains(searchTerm, "@") {
		var maintainers []models.Maintainer
		database.DBCon.Model(&maintainers).Where("email = ?", searchTerm).Select()
		if len(maintainers) > 0 {
			http.Redirect(w, r, "/maintainer/"+searchTerm, http.StatusMovedPermanently)
			return
		}
	}

	if strings.Contains(searchTerm, "*") {
		// if the query contains wildcards
		wildcardSearchTerm := strings.ReplaceAll(searchTerm, "*", "%")
		err = database.DBCon.Model(&packages).
			WhereOr("atom LIKE ? ", wildcardSearchTerm).
			WhereOr("name LIKE ? ", wildcardSearchTerm).
			Relation("Versions").
			OrderExpr("name <-> '" + searchTerm + "'").
			Select()
	} else {
		// if the query contains no wildcards do a fuzzy search
		searchQuery := BuildSearchQuery(searchTerm)
		err = database.DBCon.Model(&packages).
			Where(searchQuery).
			WhereOr("atom LIKE ? ", ("%" + searchTerm + "%")).
			Relation("Versions").
			OrderExpr("name <-> '" + searchTerm + "'").
			Select()
	}

	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	renderPackageTemplate("search",
		"search",
		GetFuncMap(),
		getSearchData(packages, searchTerm),
		w)
}

// Search renders a template containing a list of search results
// for a given query of packages
func SearchFeed(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)
	searchTerm = strings.ReplaceAll(searchTerm, "*", "")
	searchQuery := BuildSearchQuery(searchTerm)

	var packages []models.Package
	err := database.DBCon.Model(&packages).
		Where(searchQuery).
		Relation("Versions").
		OrderExpr("name <-> '" + searchTerm + "'").
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	feeds.Packages(searchTerm, packages, w)
}

func BuildSearchQuery(searchString string) string {
	var searchClauses []string
	for _, searchTerm := range strings.Split(searchString, " ") {
		if searchTerm != "" {
			searchClauses = append(searchClauses,
				"( (category % '"+searchTerm+"') OR (name % '"+searchTerm+"') OR (atom % '"+searchTerm+"') OR (maintainers @> '[{\"Name\": \""+searchTerm+"\"}]' OR maintainers @> '[{\"Email\": \""+searchTerm+"\"}]'))")
		}
	}
	return strings.Join(searchClauses, " AND ")
}
