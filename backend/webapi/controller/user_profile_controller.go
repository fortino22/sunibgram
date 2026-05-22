package controller

import (
	"strconv"

	webapiutils "github.com/sunibgram/webapi/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/dto/_shared"
)

type UserProfileController struct {
	profileService  service.IUserProfileService
	facultyService  service.IFacultyService
	hub             service.IHub
	chatRoomService service.IChatRoomService
}

func NewUserProfileController(
	profileService service.IUserProfileService,
	facultyService service.IFacultyService,
	hub service.IHub,
	chatRoomService service.IChatRoomService,
) *UserProfileController {
	return &UserProfileController{
		profileService:  profileService,
		facultyService:  facultyService,
		hub:             hub,
		chatRoomService: chatRoomService,
	}
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
	avatarFile, _ := ctx.FormFile("avatar")
	bannerFile, _ := ctx.FormFile("banner")
	clearAvatar := ctx.FormValue("clear_avatar") == "true"
	clearBanner := ctx.FormValue("clear_banner") == "true"
	if avatarFile != nil && clearAvatar {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "cannot upload avatar and clear_avatar simultaneously"})
	}
	if bannerFile != nil && clearBanner {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "cannot upload banner and clear_banner simultaneously"})
	}
	if avatarFile != nil {
		input.Avatar = avatarFile
	} else if clearAvatar {
		input.ClearAvatar = true
	}
	if bannerFile != nil {
		input.Banner = bannerFile
	} else if clearBanner {
		input.ClearBanner = true
	}
	if bio := ctx.FormValue("bio"); bio != "" {
		input.Bio = &bio
	}
	if raw := ctx.FormValue("birthday_date"); raw != "" {
		birthdayDate, err := ParseDateField(raw, "birthday_date")
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		input.BirthdayDate = &birthdayDate
	}

	avatarChanged := input.Avatar != nil || input.ClearAvatar

	if err := c.profileService.UpdateProfile(ctx.Context(), claims.UserID, input); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	if facultyID != nil {
		if err := c.facultyService.SetUserFaculty(claims.UserID, facultyID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	// Broadcast avatar change to all rooms so other users see the update in real-time
	if avatarChanged && c.hub != nil && c.chatRoomService != nil {
		go func(userID uint) {
			fp, err := c.profileService.GetProfile(userID, nil)
			if err != nil {
				return
			}
			avatarURL := ""
			if fp.Profile != nil {
				avatarURL = BuildMediaURL(fp.Profile.AvatarObjectName, fp.Profile.AvatarBucket, c.profileService.GetURL)
			}
			payload := fiber.Map{"user_id": userID, "avatar_url": avatarURL}
			c.hub.EmitToUser(userID, "user:profile_updated", payload)
			roomIDs, _ := c.chatRoomService.GetUserRoomIDs(userID)
			for _, roomID := range roomIDs {
				c.hub.Broadcast(roomID, "user:profile_updated", payload)
			}
		}(claims.UserID)
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
	gender := ""
	birthdayDate := ""
	if fp.Profile != nil {
		bio = fp.Profile.Bio
		gender = string(fp.Profile.Gender)
		birthdayDate = FormatDate(fp.Profile.BirthdayDate)
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
		Gender:         gender,
		BirthdayDate:   birthdayDate,
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
