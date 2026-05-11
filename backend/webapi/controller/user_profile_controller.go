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
	facultyService service.IFacultyService
}

func NewUserProfileController(profileService service.IUserProfileService, facultyService service.IFacultyService) *UserProfileController {
	return &UserProfileController{profileService: profileService, facultyService: facultyService}
}

func (c *UserProfileController) GetMyProfile(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	return c.fetchAndRespond(ctx, claims.UserID, nil)
}

func (c *UserProfileController) GetProfile(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	var viewerID *uint
	if claims, ok := ctxutil.Get(ctx); ok {
		viewerID = &claims.UserID
	}
	return c.fetchAndRespond(ctx, uint(id), viewerID)
}

func (c *UserProfileController) UpdateProfile(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var facultyID *uint
	if raw := ctx.FormValue("faculty_id"); raw != "" {
		fid, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid faculty_id"})
		}
		v := uint(fid)
		if _, err := c.facultyService.GetByID(v); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "faculty not found"})
		}
		facultyID = &v
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

	if facultyID != nil {
		if err := c.facultyService.SetUserFaculty(claims.UserID, facultyID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "profile updated successfully"})
}

func (c *UserProfileController) fetchAndRespond(ctx *fiber.Ctx, userID uint, viewerID *uint) error {
	fp, err := c.profileService.GetProfile(userID, viewerID)
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
	var faculty *dto.FeedAuthorFaculty
	if fp.UserFaculty != nil {
		faculty = &dto.FeedAuthorFaculty{
			FacultyID:   fp.UserFaculty.FacultyID,
			FacultyName: fp.UserFaculty.Faculty.FacultyName,
		}
	}
	return dto.UserProfileResponse{
		UserID:         fp.User.UserID,
		Name:           fp.User.Name,
		Username:       fp.User.UserName,
		Email:          fp.User.Email,
		IsPremium:      fp.User.IsPremium,
		Bio:            bio,
		AvatarURL:      avatarURL,
		BannerURL:      bannerURL,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		IsFollowing:    fp.IsFollowing,
		Faculty:        faculty,
		CreatedAt:      utils.Wrap(fp.User.CreatedAt),
	}
}
