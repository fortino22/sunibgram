package foxhistoryrepoiface

import (
	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
)

type IFoxHistoryRepository interface {
	// RecordSummary upserts a summary history row for (userID, threadID). On
	// conflict it refreshes Body/Snapshot and touches UpdatedAt so the thread
	// sorts to the top of the user's history.
	RecordSummary(userID, threadID uint, summary, snapshot string) error

	// RecordChat appends a chat history row (question + answer).
	RecordChat(userID uint, question, answer string) error

	// ListHistory returns the user's history newest-first. q filters
	// (case-insensitive) on question/body/snapshot. typeFilter, when non-empty
	// ("summary"|"chat"), restricts to that kind. Returns (rows, total).
	ListHistory(userID uint, offset, limit int, q, typeFilter string) ([]foxhistorymodel.FoxHistory, int64, error)
}
