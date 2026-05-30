package userprofilectrl

import (
	"strconv"

	webapiutils "github.com/sunibgram/webapi/utils"

	"github.com/gofiber/fiber/v2"

	chatroomsvciface "github.com/sunibgram/domain/service/chat_room"
	facultysvciface "github.com/sunibgram/domain/service/faculty"
	sockethubsvciface "github.com/sunibgram/domain/service/socket_hub"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	facultydto "github.com/sunibgram/webapi/dto/faculty"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	userprofiledto "github.com/sunibgram/webapi/dto/user_profile"
)

type UserProfileController struct {
	profileService  userprofilesvciface.IUserProfileService
	facultyService  facultysvciface.IFacultyService
	hub             sockethubsvciface.IHub
	chatRoomService chatroomsvciface.IChatRoomService
}

func NewUserProfileController(
	profileService userprofilesvciface.IUserProfileService,
	facultyService facultysvciface.IFacultyService,
	hub sockethubsvciface.IHub,
	chatRoomService chatroomsvciface.IChatRoomService,
) *UserProfileController {
	return &UserProfileController{
		profileService:  profileService,
		facultyService:  facultyService,
		hub:             hub,
		chatRoomService: chatRoomService,
	}
}

func (c *UserProfileController) GetMyFaculty(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	uf, err := c.facultyService.GetUserFaculty(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "faculty not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(facultydto.MyFacultyResponse{
		FacultyID:   uf.Faculty.FacultyID,
		FacultyName: uf.Faculty.FacultyName,
		TotalUsers:  uf.Faculty.TotalUsers,
	})
}

func (c *UserProfileController) GetMyProfile(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	uid := claims.UserID
	return c.fetchAndRespond(ctx, uid, &uid)
}

func (c *UserProfileController) GetProfile(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
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
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var facultyID *uint
	if raw := ctx.FormValue("faculty_id"); raw != "" {
		fid, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid faculty_id"})
		}
		v := uint(fid)
		if _, err := c.facultyService.GetByID(v); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "faculty not found"})
		}
		facultyID = &v
	}

	input := userprofilesvciface.UpdateProfileInput{}

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
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "cannot upload avatar and clear_avatar simultaneously"})
	}
	if bannerFile != nil && clearBanner {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "cannot upload banner and clear_banner simultaneously"})
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
		birthdayDate, err := ctrlshared.ParseDateField(raw, "birthday_date")
		if err != nil {
			return ctrlshared.HandleServiceError(ctx, err)
		}
		input.BirthdayDate = &birthdayDate
	}

	avatarChanged := input.Avatar != nil || input.ClearAvatar

	if err := c.profileService.UpdateProfile(ctx.Context(), claims.UserID, input); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	if facultyID != nil {
		if err := c.facultyService.SetUserFaculty(claims.UserID, facultyID); err != nil {
			return ctrlshared.HandleServiceError(ctx, err)
		}
	}

	// Broadcast avatar change to all rooms so other users see the update in real-time
	if avatarChanged && c.hub != nil && c.chatRoomService != nil {
		go func(userID uint) {
			defer func() { _ = recover() }()
			fp, err := c.profileService.GetProfile(userID, nil)
			if err != nil {
				return
			}
			avatarURL := ""
			if fp.Profile != nil {
				avatarURL = ctrlshared.BuildMediaURL(fp.Profile.AvatarObjectName, c.profileService.GetURL)
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
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "user not found"})
	}
	includeEmail := viewerID != nil && *viewerID == userID
	return ctx.Status(fiber.StatusOK).JSON(ToUserProfileResponse(fp, c.profileService.GetURL, includeEmail))
}

func ToUserProfileResponse(fp *userprofilesvciface.FullProfile, getURL func(string) string, includeEmail bool) userprofiledto.UserProfileResponse {
	// Blocked relationship (either direction): strip to a shell. Keep only
	// identity + block flags so the client can render the restricted state.
	// Self-view is never blocked against itself, so includeEmail callers are safe.
	if fp.IsBlocked || fp.IsBlockedBy {
		shellAvatar := ""
		if fp.Profile != nil {
			shellAvatar = ctrlshared.BuildMediaURL(fp.Profile.AvatarObjectName, getURL)
		}
		return userprofiledto.UserProfileResponse{
			UserID:      fp.User.UserID,
			Name:        fp.User.Name,
			Username:    fp.User.UserName,
			AvatarURL:   shellAvatar,
			IsFollowing: false,
			IsBlocked:   fp.IsBlocked,
			IsBlockedBy: fp.IsBlockedBy,
		}
	}

	avatarURL := ""
	bannerURL := ""
	if fp.Profile != nil {
		avatarURL = ctrlshared.BuildMediaURL(fp.Profile.AvatarObjectName, getURL)
		bannerURL = ctrlshared.BuildMediaURL(fp.Profile.BannerObjectName, getURL)
	}
	bio := ""
	gender := ""
	birthdayDate := ""
	if fp.Profile != nil {
		bio = fp.Profile.Bio
		gender = string(fp.Profile.Gender)
		birthdayDate = ctrlshared.FormatDate(fp.Profile.BirthdayDate)
	}
	followerCount := 0
	followingCount := 0
	if fp.Stat != nil {
		followerCount = fp.Stat.FollowerCount
		followingCount = fp.Stat.FollowingCount
	}
	var faculty *threaddto.FeedAuthorFaculty
	if fp.UserFaculty != nil {
		faculty = &threaddto.FeedAuthorFaculty{
			FacultyID:   fp.UserFaculty.FacultyID,
			FacultyName: fp.UserFaculty.Faculty.FacultyName,
		}
	}
	email := ""
	if includeEmail {
		email = fp.User.Email
	}
	return userprofiledto.UserProfileResponse{
		UserID:         fp.User.UserID,
		Name:           fp.User.Name,
		Username:       fp.User.UserName,
		Email:          email,
		IsPremium:      fp.User.IsPremium,
		Gender:         gender,
		BirthdayDate:   birthdayDate,
		Bio:            bio,
		AvatarURL:      avatarURL,
		BannerURL:      bannerURL,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		IsFollowing:    fp.IsFollowing,
		IsBlocked:      fp.IsBlocked,
		IsBlockedBy:    fp.IsBlockedBy,
		IsMuted:        fp.IsMuted,
		Faculty:        faculty,
		CreatedAt:      utils.Wrap(fp.User.CreatedAt),
	}
}
