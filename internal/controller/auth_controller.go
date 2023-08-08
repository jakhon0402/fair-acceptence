package controller

import (
	"context"
	"fajr-acceptance/internal/config"
	"fajr-acceptance/internal/database"
	"fajr-acceptance/internal/handler/apierr"
	"fajr-acceptance/internal/handler/middleware"
	"fajr-acceptance/internal/models"
	"fajr-acceptance/pkg/utils/authutil"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

type AuthController struct {
	JWTMiddleware *jwt.GinJWTMiddleware

	conf *config.Config
	db   *database.MongoDBClient
	//userStore store.UserStore
}

func NewAuthController(conf *config.Config, db *database.MongoDBClient) (*AuthController, error) {
	c := AuthController{
		conf: conf,
		db:   db,
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *AuthController) AuthMiddleware() gin.HandlerFunc {
	return c.JWTMiddleware.MiddlewareFunc()
}

func (c *AuthController) init() error {
	jwtconf := c.conf.Server.Auth.JWT
	m, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       jwtconf.Realm,
		Key:         []byte(jwtconf.Key),
		Timeout:     jwtconf.Timeout,
		MaxRefresh:  jwtconf.MaxRefresh,
		IdentityKey: authutil.IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if u, ok := data.(models.User); ok {
				return jwt.MapClaims{
					authutil.IdentityKey: u.Email,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: c.identityHandler,
		Authenticator:   c.authenticate,
		Authorizator:    c.authorize,
		Unauthorized:    c.unauthorized,
		LoginResponse:   c.loginResponse,
		RefreshResponse: c.loginResponse,
		TokenLookup:     "header: Authorization",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	})
	if err != nil {
		return err
	}
	c.JWTMiddleware = m
	return nil
}

// identityHandler extract user id from jwt claims and read user info from database.
func (c *AuthController) identityHandler(gctx *gin.Context) interface{} {
	var (
		ctx    = gctx.Request.Context()
		claims = jwt.ExtractClaims(gctx)
		user   models.User
		err    error
	)

	email, ok := claims[authutil.IdentityKey].(string)
	if !ok {
		return nil
	}

	filter := bson.D{
		{
			"email", email,
		},
	}

	err = c.db.GetCollection("users").FindOne(context.Background(), filter).Decode(&user)

	if err != nil {
		return nil
	}
	gctx.Request = gctx.Request.WithContext(authutil.WithUserContext(ctx, &user))
	return user
}

type SignInReq struct {
	Email    string `form:"email" json:"email" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// authenticate checks login request "POST /api/v1/login"
func (c *AuthController) authenticate(gctx *gin.Context) (interface{}, error) {
	var (
		//ctx = gctx.Request.Context()
		req SignInReq
	)
	if err := gctx.ShouldBind(&req); err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	var user models.User
	filter := bson.D{
		{
			"email", req.Email,
		},
	}

	err := c.db.GetCollection("users").FindOne(gctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	if err := authutil.MatchesPassword(user.Password, req.Password); err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	//if user.Disabled {
	//	return nil, jwt.ErrFailedAuthentication
	//}

	gctx.Request = gctx.Request.WithContext(authutil.WithUserContext(gctx.Request.Context(), &user))
	user.Sanitize(nil)
	return user, nil
}

func (c *AuthController) authorize(data interface{}, _ *gin.Context) bool {
	// Add authz at here if needed.
	_, ok := data.(models.User)
	return ok
}

func (c *AuthController) unauthorized(gctx *gin.Context, code int, message string) {
	err := apierr.ErrAuthenticationFail.WithMessage(message)
	err.RequestID = gctx.Writer.Header().Get(middleware.XRequestIdKey)
	gctx.AbortWithStatusJSON(code, err)
}

func (c *AuthController) loginResponse(gctx *gin.Context, _ int, token string, expire time.Time) {
	gctx.JSON(http.StatusOK, gin.H{
		"token": token,
		//"expire": expire.Format(time.RFC3339),
		"user": authutil.CurrentUser(gctx.Request.Context()),
	})
}
