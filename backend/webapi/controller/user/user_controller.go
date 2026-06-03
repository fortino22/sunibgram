package userctrl

import (
	"math"

	"github.com/gofiber/fiber/v2"

	rolemodel "github.com/sunibgram/domain/model/role"
	usersvciface "github.com/sunibgram/domain/service/user"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	userdto "github.com/sunibgram/webapi/dto/user"
)

type UserController struct {
	userService    usersvciface.IUserService
	profileService userprofilesvciface.IUserProfileService
}

func NewUserController(userService usersvciface.IUserService, profileService userprofilesvciface.IUserProfileService) *UserController {
	return &UserController{userService: userService, profileService: profileService}
}

func (c *UserController) GetAll(ctx *fiber.Ctx) error {
	page, limit := ctrlshared.ParsePagination(ctx)

	users, total, err := c.userService.GetAllPaginated(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.UserID
	}

	// Batch-load avatars and faculties so the admin table shows complete rows
	// without an N+1 — both lookups are cached behind the profile service.
	avatarByUser := make(map[uint]string, len(ids))
	if profiles, perr := c.profileService.GetProfiles(ids); perr == nil {
		for _, p := range profiles {
			if p.AvatarObjectName != "" {
				avatarByUser[p.UserID] = c.profileService.GetURL(p.AvatarObjectName)
			}
		}
	}
	facultyByUser := make(map[uint]userdto.UserFacultyMini, len(ids))
	if faculties, ferr := c.profileService.GetUserFaculties(ids); ferr == nil {
		for _, uf := range faculties {
			facultyByUser[uf.UserID] = userdto.UserFacultyMini{
				FacultyID:   uf.FacultyID,
				FacultyName: uf.Faculty.FacultyName,
			}
		}
	}

	data := make([]userdto.UserResponse, len(users))
	for i, u := range users {
		resp := userdto.UserResponse{
			UserID:    u.UserID,
			Name:      u.Name,
			UserName:  u.UserName,
			Email:     u.Email,
			RoleID:    u.RoleID,
			RoleName:  u.Role.Name,
			AvatarURL: avatarByUser[u.UserID],
			IsPremium: u.IsPremium,
			CreatedAt: utils.Wrap(u.CreatedAt),
		}
		if f, ok := facultyByUser[u.UserID]; ok {
			fac := f
			resp.Faculty = &fac
		}
		data[i] = resp
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return ctx.Status(fiber.StatusOK).JSON(userdto.PaginatedUsersResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// CountRegularUsers returns the total number of users with the default "user"
// role. Public endpoint — no auth required.
func (c *UserController) CountRegularUsers(ctx *fiber.Ctx) error {
	total, err := c.userService.CountUsersByRole(rolemodel.DefaultRoleName)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"total": total})
}

func (c *UserController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}

	fp, err := c.profileService.GetProfile(id, nil)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "user not found"})
	}

	detail := userdto.AdminUserDetailResponse{
		UserProfileResponse: userprofilectrl.ToUserProfileResponse(fp, c.profileService.GetURL, true),
	}
	// Role is admin-only context the public profile DTO doesn't carry; pull it
	// from the cached user lookup (Role is preloaded by FindByID).
	if user, uerr := c.userService.GetByID(id); uerr == nil && user != nil {
		detail.RoleID = user.RoleID
		detail.RoleName = user.Role.Name
	}

	return ctx.Status(fiber.StatusOK).JSON(detail)
}
