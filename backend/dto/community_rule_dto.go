package dto

import "github.com/sunibgram/dto/utils"

type CommunityRuleRequest struct {
	OrderNum int    `json:"order_num"`
	Text     string `json:"text"`
}

type CommunityRuleResponse struct {
	ID          uint                `json:"id"`
	CommunityID uint                `json:"community_id"`
	OrderNum    int                 `json:"order_num"`
	Text        string              `json:"text"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
	UpdatedAt   utils.FormattedTime `json:"updated_at"`
}

type CommunityRuleListResponse struct {
	Data []CommunityRuleResponse `json:"data"`
}

type CommunityRuleMiniResponse struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}
