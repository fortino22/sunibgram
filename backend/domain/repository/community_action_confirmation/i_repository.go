package communityactionconfirmationrepoiface

import (
	communityactionconfirmationmodel "github.com/sunibgram/domain/model/community_action_confirmation"
)

type ICommunityActionConfirmationRepository interface {
	Save(rec *communityactionconfirmationmodel.CommunityActionConfirmation) error
}
