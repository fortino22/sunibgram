package userrepoiface

import (
	"context"

	exploremodel "github.com/sunibgram/domain/model/explore"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
)

type IUserRepository interface {
	Save(user *usermodel.User) error
	SaveWithProfile(user *usermodel.User, profile *userprofilemodel.UserProfile) error
	// RegisterAccount atomically creates user + profile + settings + (optional)
	// faculty assignment in a single transaction. Returns a friendly
	// "email already registered" / "username already taken" on unique violation.
	RegisterAccount(user *usermodel.User, profile *userprofilemodel.UserProfile, settings *usersettingsmodel.UserSettings, facultyID *uint) error
	FindByID(id uint) (*usermodel.User, error)
	FindByIDs(ids []uint) ([]usermodel.User, error)
	FindByEmail(email string) (*usermodel.User, error)
	FindByUserName(username string) (*usermodel.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUserName(username string) (bool, error)
	FindAllPaginated(page, limit int) ([]usermodel.User, int64, error)
	UpdatePassword(userID uint, hashedPassword string) error
	UpdateInfo(userID uint, name, username string) error
	UpdatePremium(userID uint, premium bool) error
	UpdatePremiumByIDs(userIDs []uint, premium bool) error
	IsOnboarded(userID uint) (bool, error)
	CompleteOnboarding(userID uint) error
	// CountFollowablePeers counts non-admin users excluding the given user — the
	// pool of people a first-time user can follow to clear the onboarding gate.
	CountFollowablePeers(excludeUserID uint) (int64, error)
	// CountByRoleName counts users whose role matches the given role name.
	CountByRoleName(roleName string) (int64, error)
	SearchForShare(ctx context.Context, query string, excludeUserID uint) ([]exploremodel.ExploreUserResult, error)
}
