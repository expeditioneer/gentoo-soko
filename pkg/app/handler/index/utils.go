// miscellaneous utility functions used for the landing page of the application

package index

import (
	b64 "encoding/base64"
	"github.com/expeditioneer/gentoo-soko/pkg/app/handler/packages"
	"github.com/expeditioneer/gentoo-soko/pkg/app/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/database"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"github.com/go-pg/pg/v9/orm"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// getAddedPackages returns a list of a
// given number of recently added Versions
func getAddedPackages(n int) []models.Package {
	var addedPackages []models.Package
	err := database.DBCon.Model(&addedPackages).
		Order("preceding_commits DESC").
		Limit(n).
		Relation("Versions").
		Select()
	if err != nil {
		return addedPackages
	}
	return addedPackages
}

func getSearchHistoryPackages(r *http.Request) []models.Package {
	var cookie, err = r.Cookie("search_history")
	var searchedPackages []models.Package
	if err == nil {
		packagesList := getSearchHistoryFromCookie(cookie)
		err := database.DBCon.Model(&searchedPackages).
			Where(getSearchHistoryQuery(packagesList)).
			Relation("Versions").
			Select()
		if err != nil {
			return searchedPackages
		}
		return getSortedSearchHistory(packagesList, searchedPackages)
	}
	return searchedPackages
}

func getSortedSearchHistory(sortedPackagesList []string, packagesList []models.Package) []models.Package {
	var result []models.Package
	for _, gpackage := range sortedPackagesList {
		for _, gpackageObject := range packagesList {
			if gpackageObject.Atom == gpackage {
				result = append(result, gpackageObject)
			}
		}
	}
	reverse(result)
	return result
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func getSearchHistoryFromCookie(cookie *http.Cookie) []string {
	var packagesList []string
	cookieValue, err := b64.StdEncoding.DecodeString(cookie.Value)
	if err == nil {
		packagesList = strings.Split(string(cookieValue), ",")
		if len(packagesList) > 10 {
			packagesList = packagesList[len(packagesList)-10:]
		}
	}
	return packagesList
}

func getSearchHistoryQuery(packagesList []string) string {
	var queryParts []string
	for _, gpackage := range packagesList {
		queryParts = append(queryParts, "atom = '"+gpackage+"'")
	}
	return strings.Join(queryParts, " OR ")
}

// getUpdatedVersions returns a list of a
// given number of recently updated Versions
func getUpdatedVersions(n int) []*models.Version {
	var updatedVersions []*models.Version
	var updates []models.Commit
	err := database.DBCon.Model(&updates).
		Order("preceding_commits DESC").
		Limit(3*n).
		Relation("ChangedVersions", func(q *orm.Query) (*orm.Query, error) {
			return q.Limit(30 * n), nil
		}).
		Relation("ChangedVersions.Commits", func(q *orm.Query) (*orm.Query, error) {
			return q.Order("preceding_commits DESC"), nil
		}).
		Select()
	if err != nil {
		return updatedVersions
	}
	for _, commit := range updates {
		for _, changedVersion := range commit.ChangedVersions {
			changedVersion.Commits = changedVersion.Commits[:1]
		}
		updatedVersions = append(updatedVersions, commit.ChangedVersions...)
	}
	if len(updatedVersions) > n {
		updatedVersions = updatedVersions[:n]
	}
	return updatedVersions
}

// createPageData creates the data used in the template of the landing page
func createPageData(packagecount int, addedPackages []models.Package, updatedVersions []*models.Version, userPreferences models.UserPreferences) interface{} {
	return struct {
		Header          models.Header
		PackageCount    string
		AddedPackages   []models.Package
		UpdatedPackages []*models.Version
		Application     models.Application
		UserPreferences models.UserPreferences
	}{
		Header:          models.Header{Title: "", Tab: "home"},
		Application:     utils.GetApplicationData(),
		PackageCount:    formatPackageCount(packagecount),
		AddedPackages:   addedPackages,
		UpdatedPackages: updatedVersions,
		UserPreferences: userPreferences,
	}
}

// renderIndexTemplate renders all templates used for the landing page
func renderIndexTemplate(w http.ResponseWriter, pageData interface{}) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New("Show").
					Funcs(getFuncMap()).
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/packages/changedVersionRow.tmpl")).
			ParseGlob("web/templates/index/*.tmpl"))

	templates.ExecuteTemplate(w, "show.tmpl", pageData)
}

// GetFuncMap returns the FuncMap used in templates
func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"contains":        strings.Contains,
		"mkSlice":         mkSlice,
		"formatRestricts": packages.FormatRestricts,
	}
}

// formatPackageCount returns the formatted number of
// packages containing a thousands comma
func formatPackageCount(packageCount int) string {
	packages := strconv.Itoa(packageCount)
	if len(string(rune(packageCount))) == 6 {
		return packages[:3] + "," + packages[3:]
	} else if len(packages) == 5 {
		return packages[:2] + "," + packages[2:]
	} else if len(packages) == 4 {
		return packages[:1] + "," + packages[1:]
	} else {
		return packages
	}
}

// mkSlice creates a slice based on the given arguments
func mkSlice(args ...interface{}) []interface{} {
	return args
}
