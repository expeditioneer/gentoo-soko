// Contains functions to import categories into the database

package repository

import (
	"encoding/xml"
	"github.com/expeditioneer/gentoo-soko/internal/config"
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/logger"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"io"
	"os"
	"regexp"
	"strings"
)

// isCategory checks whether the path points to a category
// descriptions that is an metadata.xml file
func isCategory(path string) bool {
	isCategory, _ := regexp.MatchString(`[^/]*\/metadata\.xml`, path)
	return isCategory
}

// UpdateCategory updates the category in the database in case
// the given path points to a category description
func UpdateCategory(path string) {

	splittedLine := strings.Split(path, "\t")

	if len(splittedLine) != 2 {
		if len(splittedLine) == 1 && isCategory(path) {
			updateModifiedCategory(path)
		}
		return
	}

	status := splittedLine[0]
	changedFile := splittedLine[1]

	if isCategory(changedFile) && status == "D" {
		updateDeletedCategory(changedFile)
	} else if isCategory(changedFile) && (status == "A" || status == "M") {
		updateModifiedCategory(changedFile)
	}
}

// updateDeletedCategory deletes a category from the database
func updateDeletedCategory(changedFile string) {
	splitted := strings.Split(changedFile, "/")
	id := splitted[0]

	category := &models.Category{Name: id}
	_, err := database.DBCon.Model(category).WherePK().Delete()

	if err != nil {
		logger.Error.Println("Error during deleting category " + id)
		logger.Error.Println(err)
	}

}

// updateModifiedCategory adds a category to the database or
// updates it. To do so, it parses the metadata from metadata.xml
func updateModifiedCategory(changedFile string) {
	splitted := strings.Split(changedFile, "/")
	id := splitted[0]

	catmetadata := GetCatMetadata(config.PortDir() + "/" + changedFile)
	description := ""

	for _, longdescription := range catmetadata.Longdescriptions {
		if longdescription.Lang == "en" {
			description = strings.TrimSpace(longdescription.Content)
		}
	}

	category := &models.Category{
		Name:        id,
		Description: description,
	}

	_, err := database.DBCon.Model(category).OnConflict("(name) DO UPDATE").Insert()

	if err != nil {
		logger.Error.Println("Error during updating category " + id)
		logger.Error.Println(err)
	}
}

// GetCatMetadata reads and parses the category
// metadata from the metadata.xml file
func GetCatMetadata(path string) Catmetadata {
	xmlFile, err := os.Open(path)
	if err != nil {
		logger.Error.Println("Error during reading category metadata")
		logger.Error.Println(err)
	}
	defer xmlFile.Close()
	byteValue, _ := io.ReadAll(xmlFile)
	var catmetadata Catmetadata
	xml.Unmarshal(byteValue, &catmetadata)
	return catmetadata
}

// Descriptions of the category metadata.xml format

type Catmetadata struct {
	XMLName          xml.Name          `xml:"catmetadata"`
	Longdescriptions []Longdescription `xml:"longdescription"`
}

type Longdescription struct {
	XMLName xml.Name `xml:"longdescription"`
	Lang    string   `xml:"lang,attr"`
	Content string   `xml:",chardata"`
}
