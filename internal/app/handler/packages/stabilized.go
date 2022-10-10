// Used to show recently stabilized versions

package packages

import (
	"github.com/expeditioneer/gentoo-soko/internal/app/handler/feeds"
	"net/http"
)

// Stabilized renders a template containing
// a list of 50 recently stabilized versions
func Stabilized(w http.ResponseWriter, r *http.Request) {
	stabilizedVersions := GetStabilizedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Stabilized", stabilizedVersions), w)
}

func StabilizedFeed(w http.ResponseWriter, r *http.Request) {
	stabilizedVersions := GetStabilizedVersions(250)
	feeds.Changes("Stabilized packages in Gentoo.", "Stabilized packages in Gentoo.", stabilizedVersions, w)
}
