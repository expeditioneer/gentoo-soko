// Contains functions to import package mask entries into the database
//
// Example
//
// ## # Dev E. Loper <developer@gentoo.org> (2019-07-01)
// ## # Masking  these versions until we can get the
// ## # v4l stuff to work properly again
// ## =media-video/mplayer-0.90_pre5
// ## =media-video/mplayer-0.90_pre5-r1
//

package repository

import (
	"github.com/expeditioneer/gentoo-soko/internal/logger"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"github.com/expeditioneer/gentoo-soko/internal/portage/utils"
	"github.com/expeditioneer/gentoo-soko/internal/selfcheck/storage"
	"regexp"
	"strings"
	"time"
)

// isMask checks whether the path
// points to a package.mask file
func isMask(path string) bool {
	return path == "profiles/package.mask"
}

// UpdateMask updates all entries in
// the Mask table in the database
func UpdateMask(path string) {
	if isMask(path) {
		for _, packageMask := range getMasks(path) {
			parsePackageMask(packageMask)
		}
	}
}

// versionSpecifierToPackageAtom returns the package atom from a given version specifier
func versionSpecifierToPackageAtom(versionSpecifier string) string {
	gpackage := strings.ReplaceAll(versionSpecifier, ">", "")
	gpackage = strings.ReplaceAll(gpackage, "<", "")
	gpackage = strings.ReplaceAll(gpackage, "=", "")
	gpackage = strings.ReplaceAll(gpackage, "~", "")

	gpackage = strings.Split(gpackage, ":")[0]

	versionnumber := regexp.MustCompile(`-[0-9]`)
	gpackage = versionnumber.Split(gpackage, 2)[0]

	return gpackage
}

// parseAuthorLine parses the first line in the package.mask file
// and returns the author name, author email and the date
func parseAuthorLine(authorLine string) (string, string, time.Time) {

	if !(strings.Contains(authorLine, "<") && strings.Contains(authorLine, ">")) {
		logger.Error.Println("Error while parsing the author line in mask entry:")
		logger.Error.Println(authorLine)
		return "", "", time.Now()
	}

	author := strings.TrimSpace(strings.Split(authorLine, "<")[0])
	author = strings.ReplaceAll(author, "#", "")
	authorEmail := strings.TrimSpace(strings.Split(strings.Split(authorLine, "<")[1], ">")[0])
	date := strings.TrimSpace(strings.Split(authorLine, ">")[1])
	date = strings.ReplaceAll(date, "(", "")
	date = strings.ReplaceAll(date, ")", "")
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		logger.Error.Println("Error while parsing package mask date: " + date)
		logger.Error.Println(err)
	}
	return author, authorEmail, parsedDate
}

// parse the package.mask entries and
// update the Mask table in the database
func parsePackageMask(packageMask string) {
	packageMaskLines := strings.Split(packageMask, "\n")
	if len(packageMaskLines) >= 3 {
		packageMaskLine, packageMaskLines := packageMaskLines[0], packageMaskLines[1:]
		author, authorEmail, date := parseAuthorLine(packageMaskLine)

		reason := ""
		packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		for strings.HasPrefix(packageMaskLine, "#") {
			reason = reason + " " + strings.Replace(packageMaskLine, "# ", "", 1)
			packageMaskLine, packageMaskLines = packageMaskLines[0], packageMaskLines[1:]
		}

		packageMaskLines = append(packageMaskLines, packageMaskLine)

		for _, version := range packageMaskLines {
			useflag := &models.Mask{
				Author:      author,
				AuthorEmail: authorEmail,
				Date:        date,
				Reason:      reason,
				Versions:    version,
			}

			storage.Masks = append(storage.Masks, useflag)

		}
	}

}

// get all mask entries from the package.mask file
func getMasks(path string) []string {
	var masks []string
	lines, err := utils.ReadLines(path)

	if err != nil {
		logger.Error.Println("Could not read Masks file. Abort masks import")
		logger.Error.Println(err)
		return masks
	}

	line, lines := lines[0], lines[1:]
	for !strings.Contains(line, "#--- END OF EXAMPLES ---") {
		line, lines = lines[0], lines[1:]
	}
	lines = lines[1:]

	return strings.Split(strings.Join(lines, "\n"), "\n\n")
}
