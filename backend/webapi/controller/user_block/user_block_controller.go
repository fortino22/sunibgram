package userblockctrl

import (
	"github.com/gofiber/fiber/v2"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userblocksvciface "github.com/sunibgram/domain/service/user_block"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	userblockdto "github.com/sunibgram/webapi/dto/user_block"
)

type UserBlockController struct {
	blockService   userblocksvciface.IUserBlockService
	profileService userprofilesvciface.IUserProfileService
}

func NewUserBlockController(blockService userblocksvciface.IUserBlockService, profileService userprofilesvciface.IUserProfileService) *UserBlockController {
	return &UserBlockController{blockService: blockService, profileService: profileService}
}

func (c *UserBlockController) Block(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.blockService.Block(claims.UserID, targetID); err != nil {
		if err.Error() == "user not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "blocked successfully"})
}

func (c *UserBlockController) Unblock(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.blockService.Unblock(claims.UserID, targetID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unblocked successfully"})
}

func (c *UserBlockController) GetMyBlockedUsers(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	users, total, err := c.blockService.GetBlockedUsers(claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(userblockdto.PaginatedBlockedUsersResponse{
			Data: []userblockdto.BlockedUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.blockedUserProfileMap(users)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(userblockdto.PaginatedBlockedUsersResponse{
		Data: c.toBlockedUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *UserBlockController) toBlockedUserResponses(users []usermodel.User, profileMap map[uint]userprofilemodel.UserProfile) []userblockdto.BlockedUserResponse {
	data := make([]userblockdto.BlockedUserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		if profile, ok := profileMap[user.UserID]; ok {
			avatarURL = ctrlshared.BuildMediaURL(profile.AvatarObjectName, c.profileService.GetURL)
		}
		data[i] = userblockdto.BlockedUserResponse{
			UserID:    user.UserID,
			Name:      user.Name,
			UserName:  user.UserName,
			IsPremium: user.IsPremium,
			AvatarURL: avatarURL,
		}
	}
	return data
}

func (c *UserBlockController) blockedUserProfileMap(users []usermodel.User) (map[uint]userprofilemodel.UserProfile, error) {
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
