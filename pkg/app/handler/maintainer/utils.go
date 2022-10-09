package maintainer

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/expeditioneer/gentoo-soko/pkg/app/handler/packages"
	"github.com/expeditioneer/gentoo-soko/pkg/app/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

// RenderPackageTemplates renders the given templates using the given data
// One pattern can be used to specify templates
func renderMaintainerTemplate(page string, templatepattern string, funcMap template.FuncMap, data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New(page).
					Funcs(funcMap).
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/maintainer/components/*.tmpl")).
			ParseGlob("web/templates/maintainer/" + templatepattern + ".tmpl"))
	templates.ExecuteTemplate(w, page+".tmpl", data)
}

// renderIndexTemplate renders all templates used for the categories section
func renderBrowseTemplate(data interface{}, w http.ResponseWriter) {
	templates := template.Must(
		template.Must(
			template.Must(
				template.New("browse").
					ParseGlob("web/templates/layout/*.tmpl")).
				ParseGlob("web/templates/maintainer/maintainersbrowseheader.tmpl")).
			ParseGlob("web/templates/maintainer/browse.tmpl"))

	templates.ExecuteTemplate(w, "browse.tmpl", data)
}

// createPackageData creates the data used in the show package template
func createMaintainerData(pageName string, maintainer *models.Maintainer, gpackages []*models.Package) interface{} {
	return struct {
		PageName    string
		Maintainer  *models.Maintainer
		Header      models.Header
		Packages    []*models.Package
		Application models.Application
	}{
		PageName:    pageName,
		Maintainer:  maintainer,
		Header:      models.Header{Title: maintainer.Name + " – ", Tab: "maintainers"},
		Packages:    gpackages,
		Application: utils.GetApplicationData(),
	}
}

// createCategoriesData creates the data used in
// the template to display a specific category
func createBrowseData(tabName string, data interface{}) interface{} {
	return struct {
		Header      models.Header
		TabName     string
		Maintainers interface{}
		Application models.Application
	}{
		Header:      models.Header{Title: "Maintainers – ", Tab: "maintainers"},
		TabName:     tabName,
		Maintainers: data,
		Application: utils.GetApplicationData(),
	}
}

func getAllBugs(packages []*models.Package) []*models.Bug {
	allBugs := make(map[string]*models.Bug)

	for _, gpackage := range packages {
		for _, bug := range gpackage.AllBugs() {
			allBugs[bug.Id] = bug
		}
	}

	var allBugsList []*models.Bug
	for _, bug := range allBugs {
		allBugsList = append(allBugsList, bug)
	}

	sort.Slice(allBugsList, func(i, j int) bool {
		return allBugsList[i].Id < allBugsList[j].Id
	})

	return allBugsList
}

// GetFuncMap returns the FuncMap used in templates
func GetFuncMap() template.FuncMap {
	return template.FuncMap{
		"contains":        strings.Contains,
		"replaceall":      strings.ReplaceAll,
		"tolower":         strings.ToLower,
		"getAllBugs":      getAllBugs,
		"formatRestricts": packages.FormatRestricts,
		"appendCommits": func(a []*models.Commit, b []*models.Commit) []*models.Commit {
			return append(a, b...)
		},
		"gravatar": func(email string) string {
			hasher := md5.Sum([]byte(email))
			hash := hex.EncodeToString(hasher[:])
			return "https://www.gravatar.com/avatar/" + hash + "?s=13&amp;d=retro"
		},
		"getReverse": func(index int, versions []*models.Version) *models.Version {
			return versions[len(versions)-1-index]
		},
		"mkSlice": func(args ...interface{}) []interface{} {
			return args
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sortCommits": func(commits []*models.Commit) []*models.Commit {
			sort.Slice(commits, func(i, j int) bool {
				return commits[i].PrecedingCommits > commits[j].PrecedingCommits
			})
			return commits
		},
		"getPullRequests": func(packages []*models.Package) []*models.GithubPullRequest {
			pullrequestsMap := map[string]*models.GithubPullRequest{}
			for _, gpackage := range packages {
				for _, pr := range gpackage.PullRequests {
					pullrequestsMap[pr.Id] = pr
				}
			}

			var pullrequests []*models.GithubPullRequest
			for _, pr := range pullrequestsMap {
				pullrequests = append(pullrequests, pr)
			}
			return pullrequests
		},
	}
}
