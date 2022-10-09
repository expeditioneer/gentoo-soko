// Contains functions to import USE flags into the database

package repository

import (
	"github.com/expeditioneer/gentoo-soko/pkg/config"
	"github.com/expeditioneer/gentoo-soko/pkg/logger"
	"github.com/expeditioneer/gentoo-soko/pkg/models"
	"github.com/expeditioneer/gentoo-soko/pkg/portage/utils"
	"github.com/expeditioneer/gentoo-soko/pkg/selfcheck/storage"
	"strings"
)

// UpdateUse reads all USE flags descriptions from the given file in
// case the given file contains USE flags descriptions and imports
// each USE flag into the database
func UpdateUse(path string) {

	if isLocalUseflag(path) || isGlobalUseflag(path) || isUseExpand(path) {

		rawFlags, _ := utils.ReadLines(config.PortDir() + "/" + path)

		for _, rawFlag := range rawFlags {

			if strings.TrimSpace(rawFlag) == "" || rawFlag[0:1] == "#" {
				continue
			}

			scope := getScope(path)

			var err error
			if scope == "local" || scope == "global" {
				err = createUseflag(rawFlag, scope)
			} else if scope == "use_expand" {
				file := strings.Split(path, "/")[2]
				err = createUseExpand(rawFlag, file)
			}

			if err != nil {
				logger.Info.Println("Error during updating useflag " + rawFlag)
				logger.Info.Println(err)
				logger.Error.Println("Error during updating useflag " + rawFlag)
				logger.Error.Println(err)
			}
		}
	}

}

// createUseflag parses the description from the file,
// creates a USE flag and imports it into the database
func createUseflag(rawFlag string, scope string) error {
	line := strings.Split(rawFlag, " - ")
	splitted := strings.Split(line[0], ":")
	gpackage := ""

	if scope == "local" {
		gpackage = splitted[0]
	}

	useflag := &models.Useflag{
		Id:          line[0] + "-" + scope,
		Package:     gpackage,
		Name:        splitted[len(splitted)-1],
		Scope:       scope,
		Description: strings.Join(line[1:], ""),
	}

	storage.Useflags = append(storage.Useflags, useflag)

	return nil
}

// createUseExpand parses the description from the file,
// creates a USE expand flag and imports it into the database
func createUseExpand(rawFlag string, file string) error {
	name := strings.ReplaceAll(file, ".desc", "")
	line := strings.Split(rawFlag, " - ")
	id := name + "_" + line[0]

	useExpand := &models.Useflag{
		Id:          id,
		Name:        name + "_" + line[0],
		Scope:       "use_expand",
		Description: strings.Join(line[1:], ""),
		UseExpand:   name,
	}

	storage.Useflags = append(storage.Useflags, useExpand)

	return nil
}

// getScope returns either "local", "global", "use_expand"
// or "" based on the file that the path points to
func getScope(path string) string {
	if isLocalUseflag(path) {
		return "local"
	} else if isGlobalUseflag(path) {
		return "global"
	} else if isUseExpand(path) {
		return "use_expand"
	}
	return ""
}

// isLocalUseflag checks whether the path points to
// the file that contains the local USE flags
func isLocalUseflag(path string) bool {
	return path == "profiles/use.local.desc"
}

// isGlobalUseflag checks whether the path points to
// the file that contains the global USE flags
func isGlobalUseflag(path string) bool {
	return path == "profiles/use.desc"
}

// isGlobalUseflag checks whether the path points to
// a file that contains use expand flags
func isUseExpand(path string) bool {
	return strings.HasPrefix(path, "profiles/desc/")
}
