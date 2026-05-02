package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type UserProfileController struct {
	profileService service.IUserProfileService
}

func NewUserProfileController(profileService service.IUserProfileService) *UserProfileController {
	return &UserProfileController{profileService: profileService}
}

func (c *UserProfileController) GetMyProfile(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	return c.fetchAndRespond(ctx, claims.UserID)
}

func (c *UserProfileController) GetProfile(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	return c.fetchAndRespond(ctx, uint(id))
}

func (c *UserProfileController) UpdateProfile(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	input := service.UpdateProfileInput{}

	if name := ctx.FormValue("name"); name != "" {
		input.Name = &name
	}
	if username := ctx.FormValue("username"); username != "" {
		input.Username = &username
	}
	if avatar, err := ctx.FormFile("avatar"); err == nil {
		input.Avatar = avatar
	} else if ctx.FormValue("clear_avatar") == "true" {
		input.ClearAvatar = true
	}
	if banner, err := ctx.FormFile("banner"); err == nil {
		input.Banner = banner
	} else if ctx.FormValue("clear_banner") == "true" {
		input.ClearBanner = true
	}
	if bio := ctx.FormValue("bio"); bio != "" {
		input.Bio = &bio
	}

	if err := c.profileService.UpdateProfile(ctx.Context(), claims.UserID, input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "profile updated successfully"})
}

func (c *UserProfileController) fetchAndRespond(ctx *fiber.Ctx, userID uint) error {
	fp, err := c.profileService.GetProfile(userID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "user not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(fp))
}

func (c *UserProfileController) toResponse(fp *service.FullProfile) dto.UserProfileResponse {
	avatarURL := ""
	if fp.Profile != nil && fp.Profile.AvatarObjectName != "" {
		avatarURL = c.profileService.GetURL(fp.Profile.AvatarBucket, fp.Profile.AvatarObjectName)
	}
	bannerURL := ""
	if fp.Profile != nil && fp.Profile.BannerObjectName != "" {
		bannerURL = c.profileService.GetURL(fp.Profile.BannerBucket, fp.Profile.BannerObjectName)
	}
	bio := ""
	if fp.Profile != nil {
		bio = fp.Profile.Bio
	}
	followerCount := 0
	followingCount := 0
	if fp.Stat != nil {
		followerCount = fp.Stat.FollowerCount
		followingCount = fp.Stat.FollowingCount
	}
	return dto.UserProfileResponse{
		UserID:         fp.User.UserID,
		Name:           fp.User.Name,
		Username:       fp.User.UserName,
		Email:          fp.User.Email,
		Bio:            bio,
		AvatarURL:      avatarURL,
		BannerURL:      bannerURL,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		CreatedAt:      utils.Wrap(fp.User.CreatedAt),
	}
}
