package controller

import (
	"math"

	"github.com/gofiber/fiber/v2"

	usersvciface "github.com/sunibgram/domain/service/user"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
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
	page, limit := ParsePagination(ctx)

	users, total, err := c.userService.GetAllPaginated(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	data := make([]userdto.UserResponse, len(users))
	for i, u := range users {
		data[i] = userdto.UserResponse{
			UserID:    u.UserID,
			Name:      u.Name,
			UserName:  u.UserName,
			Email:     u.Email,
			RoleID:    u.RoleID,
			RoleName:  u.Role.Name,
			CreatedAt: utils.Wrap(u.CreatedAt),
		}
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

func (c *UserController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}

	fp, err := c.profileService.GetProfile(id, nil)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "user not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(toUserProfileResponse(fp, c.profileService.GetURL))
}
