package trendreportdto

type CreateTrendReportRequest struct {
	HashtagName string `json:"hashtag_name"`
	ReasonCode  string `json:"reason_code"`
}

type TrendReportReasonResponse struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type TrendReportReasonStatsResponse struct {
	Code  string `json:"code"`
	Label string `json:"label"`
	Count int64  `json:"count"`
}
