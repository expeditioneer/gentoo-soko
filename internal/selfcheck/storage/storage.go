// Contains utility functions to execute commands and parse the output

package storage

import "github.com/expeditioneer/gentoo-soko/internal/models"

var (
	Packages   []*models.Package
	Versions   []*models.Version
	Useflags   []*models.Useflag
	Masks      []*models.Mask
	Categories []*models.Category
)
