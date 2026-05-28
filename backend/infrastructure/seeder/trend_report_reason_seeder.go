package seeder

import (
	"log"

	"gorm.io/gorm"

	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
)

var trendReportReasonSeeds = []trendreportmodel.TrendReportReason{
	{Code: "not_relevant", Label: "The associated content is not relevant"},
	{Code: "spam", Label: "This trend is spam"},
	{Code: "abusive_or_harmful", Label: "This trend is abusive or harmful"},
	{Code: "not_interested", Label: "Not interested in this"},
	{Code: "duplicate", Label: "This trend is a duplicate"},
	{Code: "harmful_or_spammy", Label: "This trend is harmful or spammy"},
}

func SeedTrendReportReason(db *gorm.DB) {
	for _, seed := range trendReportReasonSeeds {
		reason := seed
		if err := db.Where(trendreportmodel.TrendReportReason{Code: seed.Code}).FirstOrCreate(&reason).Error; err != nil {
			log.Printf("trend report reason seeder: failed to seed %s: %v", seed.Code, err)
			continue
		}
	}
}
