// miscellaneous utility functions used for the about pages

package about

import (
	"github.com/expeditioneer/gentoo-soko/pkg/app/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"html/template"
	"net/http"
)

// renderAboutTemplate renders a specific about template
func renderAboutTemplate(w http.ResponseWriter, r *http.Request, page string) {
	templates := template.Must(
		template.Must(
			template.New(page).
				ParseGlob("web/templates/layout/*.tmpl")).
			ParseGlob("web/templates/about/" + page + ".tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", getPageData())
}

// getPageData returns the data used
// in all about templates
func getPageData() interface{} {
	return struct {
		Header      models.Header
		Application models.Application
	}{
		Header:      models.Header{Title: "About – ", Tab: "about"},
		Application: utils.GetApplicationData(),
	}
}
