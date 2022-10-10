package selfcheck

import (
	"github.com/expeditioneer/gentoo-soko/internal/database"
	"github.com/expeditioneer/gentoo-soko/internal/logger"
	"github.com/expeditioneer/gentoo-soko/internal/models"
	"github.com/expeditioneer/gentoo-soko/internal/selfcheck/metrics"
	"github.com/expeditioneer/gentoo-soko/internal/selfcheck/repository"
	"github.com/expeditioneer/gentoo-soko/internal/selfcheck/storage"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func AllPackages() {
	logger.Info.Println("selfcheck: Preparing new check...")
	logger.Info.Println("selfcheck: Updating selfcheck repository")
	repository.UpdateRepo()
	logger.Info.Println("selfcheck: Importing data")
	repository.Import()
	logger.Info.Println("selfcheck: Resetting metrics")
	resetMetrics()

	logger.Info.Println("selfcheck: Start check")
	for _, category := range storage.Categories {
		//logger.Info.Println("Checking " + category.Name)
		checkCategory(category)
	}
	logger.Info.Println("selfcheck: Finished check")
}

func resetMetrics() {
	for _, metric := range metrics.MissingPackages {
		prometheus.Unregister(metric)
	}
	for _, metric := range metrics.MissingVersions {
		prometheus.Unregister(metric)
	}
	metrics.MissingPackages = map[string]prometheus.Gauge{}
	metrics.MissingVersions = map[string]prometheus.Gauge{}
}

func checkCategory(category *models.Category) {
	// create a client (safe to share across requests)

	database.Connect()
	defer database.DBCon.Close()

	pgoCategory := new(models.Category)
	err := database.DBCon.Model(pgoCategory).
		Where("name = ?", category.Name).
		Relation("Packages").
		Relation("Packages.Versions").
		Select()

	if err != nil {
		logger.Error.Println(err)
		return
	}

	for _, localPackage := range storage.Packages {
		if localPackage.Category == category.Name {
			var matchingRemotePackage *models.Package
			for _, remotePackage := range pgoCategory.Packages {
				if localPackage.Atom == remotePackage.Atom {
					matchingRemotePackage = remotePackage
					break
				}
			}

			if matchingRemotePackage == nil {
				// register outdated
				if metric, ok := metrics.MissingPackages[localPackage.Atom]; ok {
					metric.Set(1)
				} else {
					metrics.MissingPackages[localPackage.Atom] = promauto.NewGauge(prometheus.GaugeOpts{
						Name:        "pgo_missing_package",
						Help:        "A package that is missing on packages.g.o although it's present in the tree",
						ConstLabels: prometheus.Labels{"atom": localPackage.Atom},
					})
					metrics.MissingPackages[localPackage.Atom].Set(1)
				}
			} else {
				checkVersions(matchingRemotePackage)
			}
		}
	}

}

func checkVersions(remotePackage *models.Package) {

	for _, localVersion := range storage.Versions {

		if localVersion.Atom == remotePackage.Atom {

			// search for local version in remote versions
			versionFound := false
			for _, remoteVersion := range remotePackage.Versions {
				if localVersion.Id == remoteVersion.Id {
					versionFound = true
					break
				}
			}

			if !versionFound {
				if metric, ok := metrics.MissingVersions[localVersion.Id]; ok {
					metric.Set(1)
				} else {
					metrics.MissingVersions[localVersion.Id] = promauto.NewGauge(prometheus.GaugeOpts{
						Name:        "pgo_missing_version",
						Help:        "A version that is missing on packages.g.o although it's present in the tree",
						ConstLabels: prometheus.Labels{"id": localVersion.Id},
					})
					metrics.MissingVersions[localVersion.Id].Set(1)
				}
			}

			// TODO: check mask entries

		}
	}
}
