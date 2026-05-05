package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func Register(
	app *fiber.App,
	authCtrl *controller.AuthController,
	roleCtrl *controller.RoleController,
	userCtrl *controller.UserController,
	profileCtrl *controller.UserProfileController,
	threadCtrl *controller.ThreadController,
	likeCtrl *controller.LikeController,
	bookmarkCtrl *controller.BookmarkController,
	repostCtrl *controller.RepostController,
	emailCtrl *controller.EmailController,
	verificationCtrl *controller.VerificationController,
	followCtrl *controller.FollowController,
	communityCtrl *controller.CommunityController,
	communityMemberCtrl *controller.CommunityMemberController,
	communityRequestCtrl *controller.CommunityRequestController,
	communityThreadCtrl *controller.CommunityThreadController,
	facultyCtrl *controller.FacultyController,
	orderCtrl *controller.OrderController,
) {
	api := app.Group("/api")
	auth := middleware.Auth()

	registerAuthRoutes(api, authCtrl, verificationCtrl, auth)
	registerVerificationRoutes(api, verificationCtrl)
	registerProfileRoutes(api, profileCtrl, auth)
	registerThreadRoutes(api, threadCtrl, likeCtrl, bookmarkCtrl, repostCtrl, auth)
	registerEmailRoutes(api, emailCtrl, auth)
	registerUserRoutes(api, followCtrl, auth)
	registerAdminRoutes(api, roleCtrl, userCtrl)
	registerCommunityRoutes(api, communityCtrl, communityMemberCtrl, communityRequestCtrl, communityThreadCtrl, auth)
	registerFacultyRoutes(api, facultyCtrl, auth)
	registerOrderRoutes(api, orderCtrl, auth)
}
