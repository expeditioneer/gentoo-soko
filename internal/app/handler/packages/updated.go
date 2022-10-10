// Used to show recently updated versions

package packages

import (
	"github.com/expeditioneer/gentoo-soko/internal/app/handler/feeds"
	"net/http"
)

// Updated renders a template containing
// a list of 50 recently updated versions
func Updated(w http.ResponseWriter, r *http.Request) {
	updatedVersions := GetUpdatedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Updated", updatedVersions), w)
}

func UpdatedFeed(w http.ResponseWriter, r *http.Request) {
	updatedVersions := GetUpdatedVersions(250)
	feeds.Changes("Added packages in Gentoo.", "Added packages in Gentoo.", updatedVersions, w)
}
