package arches

import (
	"net/http"
	utils2 "soko/pkg/app/utils"
)

// Index renders a template to show the landing page containing links to all arches feeds
func Index(w http.ResponseWriter, r *http.Request) {
	userPreferences := utils2.GetUserPreferences(r)
	http.Redirect(w, r, "/arches/"+userPreferences.Arches.DefaultArch+"/"+userPreferences.Arches.DefaultPage, http.StatusSeeOther)
}
