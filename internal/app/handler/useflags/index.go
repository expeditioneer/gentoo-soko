// Used to display the landing page of the USE flag section

package useflags

import (
	utils2 "github.com/expeditioneer/gentoo-soko/internal/app/utils"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"html/template"
	"net/http"
)

// Index renders a template to show the index page of the USE flags
// section containing a bubble chart of popular USE flags
func Index(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Header      models.Header
		Page        string
		Application models.Application
	}{
		Header:      models.Header{Title: "Useflags – ", Tab: "useflags"},
		Page:        "browse",
		Application: utils2.GetApplicationData(),
	}

	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/useflags/browseuseflagsheader.tmpl")).
			ParseGlob("web/templates/useflags/index.tmpl"))

	templates.ExecuteTemplate(w, "index.tmpl", data)
}

// Index renders a template to show the index page of the USE flags
// section containing a bubble chart of popular USE flags
func Default(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils2.GetUserPreferences(r)
	if userPreferences.Useflags.Layout == "bubble" {
		http.Redirect(w, r, "/useflags/popular", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/useflags/search", http.StatusSeeOther)
	}
}
