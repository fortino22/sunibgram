package controller

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type UserController struct {
	userService service.IUserService
}

func NewUserController(userService service.IUserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) GetAll(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	users, total, err := c.userService.GetAllPaginated(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.UserResponse, len(users))
	for i, u := range users {
		data[i] = dto.UserResponse{
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

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedUsersResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

func (c *UserController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	user, err := c.userService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "user not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.UserResponse{
		UserID:    user.UserID,
		Name:      user.Name,
		UserName:  user.UserName,
		Email:     user.Email,
		RoleID:    user.RoleID,
		RoleName:  user.Role.Name,
		CreatedAt: utils.Wrap(user.CreatedAt),
	})
}

func (c *UserController) GetProfile(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	user, stat, err := c.userService.GetProfile(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "user not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.UserProfileResponse{
		UserID:         user.UserID,
		Name:           user.Name,
		Username:       user.UserName,
		Email:          user.Email,
		FollowerCount:  stat.FollowerCount,
		FollowingCount: stat.FollowingCount,
		CreatedAt:      utils.Wrap(user.CreatedAt),
	})
}

func (c *UserController) GetMyProfile(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	user, stat, err := c.userService.GetProfile(claims.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "user not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(dto.UserProfileResponse{
		UserID:         user.UserID,
		Name:           user.Name,
		Username:       user.UserName,
		Email:          user.Email,
		FollowerCount:  stat.FollowerCount,
		FollowingCount: stat.FollowingCount,
		CreatedAt:      utils.Wrap(user.CreatedAt),
	})
}
