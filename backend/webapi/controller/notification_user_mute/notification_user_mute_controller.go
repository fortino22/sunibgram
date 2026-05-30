package notificationusermutectrl

import (
	"github.com/gofiber/fiber/v2"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	notificationusermutesvciface "github.com/sunibgram/domain/service/notification_user_mute"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	notificationusermutedto "github.com/sunibgram/webapi/dto/notification_user_mute"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type NotificationUserMuteController struct {
	muteService    notificationusermutesvciface.INotificationUserMuteService
	profileService userprofilesvciface.IUserProfileService
}

func NewNotificationUserMuteController(muteService notificationusermutesvciface.INotificationUserMuteService, profileService userprofilesvciface.IUserProfileService) *NotificationUserMuteController {
	return &NotificationUserMuteController{muteService: muteService, profileService: profileService}
}

func (c *NotificationUserMuteController) Mute(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.muteService.Mute(claims.UserID, targetID); err != nil {
		if err.Error() == "user not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "muted successfully"})
}

func (c *NotificationUserMuteController) Unmute(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.muteService.Unmute(claims.UserID, targetID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unmuted successfully"})
}

func (c *NotificationUserMuteController) GetMyMutedUsers(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	users, total, err := c.muteService.GetMutedUsers(claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(notificationusermutedto.PaginatedMutedUsersResponse{
			Data: []notificationusermutedto.MutedUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.mutedUserProfileMap(users)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(notificationusermutedto.PaginatedMutedUsersResponse{
		Data: c.toMutedUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *NotificationUserMuteController) toMutedUserResponses(users []usermodel.User, profileMap map[uint]userprofilemodel.UserProfile) []notificationusermutedto.MutedUserResponse {
	data := make([]notificationusermutedto.MutedUserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		if profile, ok := profileMap[user.UserID]; ok {
			avatarURL = ctrlshared.BuildMediaURL(profile.AvatarObjectName, c.profileService.GetURL)
		}
		data[i] = notificationusermutedto.MutedUserResponse{
			UserID:    user.UserID,
			Name:      user.Name,
			UserName:  user.UserName,
			IsPremium: user.IsPremium,
			AvatarURL: avatarURL,
		}
	}
	return data
}

func (c *NotificationUserMuteController) mutedUserProfileMap(users []usermodel.User) (map[uint]userprofilemodel.UserProfile, error) {
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.UserID)
	}

	profiles, err := c.profileService.GetProfiles(userIDs)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[uint]userprofilemodel.UserProfile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}
	return profileMap, nil
}
