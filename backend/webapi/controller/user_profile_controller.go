package controller

import (
	webapiutils "github.com/sunibgram/webapi/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type UserProfileController struct {
	profileService service.IUserProfileService
	facultyService service.IFacultyService
}

func NewUserProfileController(profileService service.IUserProfileService, facultyService service.IFacultyService) *UserProfileController {
	return &UserProfileController{profileService: profileService, facultyService: facultyService}
}

func (c *UserProfileController) GetMyFaculty(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	uf, err := c.facultyService.GetUserFaculty(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "faculty not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.MyFacultyResponse{
		FacultyID:   uf.Faculty.FacultyID,
		FacultyName: uf.Faculty.FacultyName,
		TotalUsers:  uf.Faculty.TotalUsers,
	})
}

func (c *UserProfileController) GetMyProfile(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	return c.fetchAndRespond(ctx, claims.UserID, nil)
}

func (c *UserProfileController) GetProfile(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	var viewerID *uint
	if claims, ok := webapiutils.Get(ctx); ok {
		viewerID = &claims.UserID
	}
	return c.fetchAndRespond(ctx, id, viewerID)
}

func (c *UserProfileController) UpdateProfile(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
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
	return toUserProfileResponse(fp, c.profileService.GetURL)
}

func toUserProfileResponse(fp *service.FullProfile, getURL func(string, string) string) dto.UserProfileResponse {
	avatarURL := ""
	bannerURL := ""
	if fp.Profile != nil {
		avatarURL = BuildMediaURL(fp.Profile.AvatarObjectName, fp.Profile.AvatarBucket, getURL)
		bannerURL = BuildMediaURL(fp.Profile.BannerObjectName, fp.Profile.BannerBucket, getURL)
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
