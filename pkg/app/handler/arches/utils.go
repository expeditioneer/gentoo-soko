// miscellaneous utility functions used for arches

package arches

import (
	"github.com/expeditioneer/gentoo-soko/pkg/app/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/database"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"html/template"
	"net/http"
)

// getPageData creates the data used in all
// templates used in the arches section
func getPageData() interface{} {
	return struct {
		Header      models.Header
		Application models.Application
	}{
		Header:      models.Header{Title: "Architectures – ", Tab: "arches"},
		Application: utils.GetApplicationData(),
	}
}

// getStabilizedVersionsForArch returns the given number of recently
// stabilized versions of a specific arch
func getStabilizedVersionsForArch(arch string, n int) ([]*models.Version, error) {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("stabilized::jsonb @> ?", "\""+arch+"\"").
		Limit(n).
		Select()
	if err != nil {
		return nil, err
	}

	for _, update := range updates {
		if update.Version != nil {
			update.Version.Commits = []*models.Commit{update.Commit}
			stabilizedVersions = append(stabilizedVersions, update.Version)
		}
	}

	return stabilizedVersions, err
}

// getKeywordedVersionsForArch returns the given number of recently
// keyworded versions of a specific arch
func getKeywordedVersionsForArch(arch string, n int) ([]*models.Version, error) {
	var stabilizedVersions []*models.Version
	var updates []models.KeywordChange
	err := database.DBCon.Model(&updates).
		Relation("Version").
		Relation("Commit").
		Order("commit.preceding_commits DESC").
		Where("added::jsonb @> ?", "\""+arch+"\"").
		Limit(n).
		Select()
	if err != nil {
		return nil, err
	}

	for _, update := range updates {
		if update.Version != nil {
			update.Version.Commits = []*models.Commit{update.Commit}
			stabilizedVersions = append(stabilizedVersions, update.Version)
		}
	}

	return stabilizedVersions, err
}

// RenderPackageTemplates renders the arches templates using the given data
func renderPackageTemplates(page string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter) {

	templates := template.Must(
		template.Must(
			template.Must(
				template.Must(
					template.New(page).
						Funcs(funcMap).
						ParseGlob("web/templates/layout/*.tmpl")).
					ParseGlob("web/templates/arches/archesheader.tmpl")).
				ParseGlob("web/templates/arches/changedVersionRows.tmpl")).
			ParseGlob("web/templates/arches/changedVersions.tmpl"))

	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// CreateFeedData creates the data used in changedVersions template
func createFeedData(arch string, name string, feedtype string, versions []*models.Version, userPreferences models.UserPreferences) interface{} {
	return struct {
		Header          models.Header
		Arch            string
		Name            string
		FeedName        string
		Versions        []*models.Version
		Application     models.Application
		UserPreferences models.UserPreferences
	}{
		Header:          models.Header{Title: "Architectures – ", Tab: "arches"},
		Arch:            arch,
		Name:            name,
		FeedName:        feedtype,
		Versions:        versions,
		Application:     utils.GetApplicationData(),
		UserPreferences: userPreferences,
	}
}
