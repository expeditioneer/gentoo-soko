// Used to search for USE flags

package useflags

import (
	utils2 "github.com/expeditioneer/gentoo-soko/internal/app/utils"
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"github.com/go-pg/pg"
	"html/template"
	"net/http"
)

// Search renders a template containing a list of search results
// for a given query of USE flags
func Global(w http.ResponseWriter, r *http.Request) {

	var useflags []models.Useflag
	err := database.DBCon.Model(&useflags).Where("scope = 'global'").Select()
	if err != nil && err != pg.ErrNoRows {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	data := struct {
		Header      models.Header
		Page        string
		Useflags    []models.Useflag
		Application models.Application
	}{
		Header:      models.Header{Title: "Global" + " – ", Tab: "useflags"},
		Page:        "global",
		Useflags:    useflags,
		Application: utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/browseuseflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/list.tmpl"))

	templates.ExecuteTemplate(w, "list.tmpl", data)
}
