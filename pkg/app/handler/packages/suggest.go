// Used to create package suggestions

package packages

import (
	"encoding/json"
	"github.com/expeditioneer/gentoo-soko/pkg/database"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"github.com/go-pg/pg"
	"net/http"
)

// Suggest returns json encoded suggestions of
// packages based on the given query
func Suggest(w http.ResponseWriter, r *http.Request) {

	searchTerm := getParameterValue("q", r)

	var packages []models.Package
	err := database.DBCon.Model(&packages).
		Where("atom LIKE ? ", ("%" + searchTerm + "%")).
		Relation("Versions").
		Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	type Result struct {
		Name        string `json:"name"`
		Category    string `json:"category"`
		description string `json:"description"`
	}

	type Results struct {
		Results []*Result `json:"results"`
	}

	var results []*Result

	for _, gpackage := range packages {
		results = append(results, &Result{
			Name:        gpackage.Name,
			Category:    gpackage.Category,
			description: gpackage.Versions[0].Description,
		})
	}

	result := Results{
		Results: results,
	}

	b, err := json.Marshal(result)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
