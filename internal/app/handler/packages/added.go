// Used to show recently added versions

package packages

import (
	"github.com/expeditioneer/gentoo-soko/internal/app/handler/feeds"
	"net/http"
)

// Added renders a template containing a list of 50 recently added versions.
func Added(w http.ResponseWriter, r *http.Request) {
	addedVersions := GetAddedVersions(50)
	RenderPackageTemplates("changedVersions", "changedVersions", "changedVersionRow", GetFuncMap(), CreateFeedData("Added", addedVersions), w)
}

func AddedFeed(w http.ResponseWriter, r *http.Request) {
	addedPackages := GetAddedPackages(250)
	feeds.AddedPackages("Added packages in Gentoo.", "Added packages in Gentoo.", addedPackages, w)
}
