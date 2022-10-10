package repology

import (
	"bytes"
	"encoding/json"
	"github.com/expeditioneer/gentoo-soko/internal/config"
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/logger"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Package struct {
	Repo        string `json:"repo"`
	Name        string `json:"name"`
	VisibleName string `json:"visiblename"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

type Packages map[string][]Package

// UpdateOutdated will update the database table that contains all outdated gentoo versions
func UpdateOutdated() {

	database.Connect()
	defer database.DBCon.Close()

	if config.Quiet() == "true" {
		log.SetOutput(io.Discard)
	}

	// Get all outdated Versions
	var outdatedVersions []*models.OutdatedPackages
	letters := "abcdefghijklmnopqrstuvwxyz"
	for _, letter := range letters {
		outdatedVersions = append(outdatedVersions, getOutdatedStartingWith(letter)...)
	}

	// Clean up the database
	deleteAllOutdated()

	// Update the database
	for _, outdated := range outdatedVersions {
		database.DBCon.Insert(outdated)
	}

	updateStatus()
}

// getOutdatedStartingWith gets all outdated packages starting with the given letter
func getOutdatedStartingWith(letter rune) []*models.OutdatedPackages {
	repoPackages, err := parseRepologyData("https://repology.org/api/v1/projects/" + string(letter) + "/?inrepo=gentoo&outdated=1")

	if err != nil {
		logger.Error.Println("Error while fetching repology data")
		return []*models.OutdatedPackages{}
	}

	blockedRepos := readBlocklist("ignored-repositories")
	blockedCategories := readBlocklist("ignored-categories")
	blockedPackages := readBlocklist("ignored-packages")

	var outdatedVersions []*models.OutdatedPackages
	for packagename := range repoPackages {
		atom := ""
		newest := ""
		version := ""
		outdated := false
		// get the gentoo atom name first
		for _, v := range repoPackages[packagename] {
			if v.Repo == "gentoo" {
				atom = v.VisibleName
			}
		}
		for _, v := range repoPackages[packagename] {
			if v.Status == "newest" &&
				!contains(blockedRepos, v.Repo) &&
				!contains(blockedPackages, atom+"::"+v.Repo) {
				newest = v.Version
			}
			if v.Repo == "gentoo" && v.Status == "newest" {
				outdated = false
				break
			}
			if v.Repo == "gentoo" &&
				v.Status == "outdated" &&
				!containsPrefix(blockedCategories, strings.Split(v.VisibleName, "/")[0]) &&
				!containsPrefix(blockedPackages, v.VisibleName) {

				atom = v.VisibleName
				outdated = true
				version = v.Version
			}
		}

		if outdated && newest != "" && strings.HasPrefix(packagename, string(letter)) {
			outdatedVersions = append(outdatedVersions, &models.OutdatedPackages{
				Atom:          atom,
				GentooVersion: version,
				NewestVersion: newest,
			})
		}
	}

	return outdatedVersions
}

// parseRepologyData gets the json from given url and parses it
func parseRepologyData(url string) (Packages, error) {
	resp, err := http.Get(url)
	if err != nil {
		return Packages{}, err
	}
	defer resp.Body.Close()
	var repoPackages Packages
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&repoPackages)
	return repoPackages, err
}

// deleteAllOutdated deletes all entries in the outdated table
func deleteAllOutdated() {
	var allOutdated []*models.OutdatedPackages
	database.DBCon.Model(&allOutdated).Select()
	for _, outdated := range allOutdated {
		database.DBCon.Model(outdated).WherePK().Delete()
	}
}

// readBlocklist parses a block list and returns a list of
// lines whereas comments as well as empty lines are ignored
func readBlocklist(file string) []string {
	var blocklist []string
	resp, err := http.Get("https://gitweb.gentoo.org/sites/soko-metadata.git/plain/repology/" + file)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	rawBlocklist := buf.String()

	for _, line := range strings.Split(rawBlocklist, "\n") {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			blocklist = append(blocklist, line)
		}
	}
	return blocklist
}

// contains returns true if the given list includes
// the given string. Otherwise false is returned.
func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}

// contains returns true if the given string is a prefix
// of an item in the given list. Otherwise false is returned.
func containsPrefix(list []string, item string) bool {
	for _, i := range list {
		if strings.HasPrefix(i, item) {
			return true
		}
	}
	return false
}

func updateStatus() {
	database.DBCon.Model(&models.Application{
		Id:         "repology",
		LastUpdate: time.Now(),
		Version:    config.Version(),
	}).OnConflict("(id) DO UPDATE").Insert()
}
