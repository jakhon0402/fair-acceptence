package server

import (
	"context"
	"errors"
	"fajr-acceptance/internal/config"
	"fajr-acceptance/internal/controller"
	"fajr-acceptance/internal/database"
	"fajr-acceptance/internal/handler"
	"fajr-acceptance/internal/handler/middleware"
	"fajr-acceptance/pkg/logger"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"go.uber.org/multierr"
	"net/http"
	"sync/atomic"
)

type Server struct {
	apiServer        *http.Server
	conf             *config.Config
	running          int32
	apiEngine        *gin.Engine
	authController   *controller.AuthController
	courseController *controller.CourseController
}

func NewServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger *logrus.Logger,
	db *database.MongoDBClient,
	authController *controller.AuthController,
	courseController *controller.CourseController,

) (*Server, error) {

	gin.SetMode(gin.DebugMode)

	srv := Server{
		conf:             cfg,
		apiEngine:        gin.Default(),
		authController:   authController,
		courseController: courseController,
	}
	corscfg := cors.DefaultConfig()
	corscfg.AllowAllOrigins = true
	corscfg.AllowCredentials = true
	corscfg.AddAllowMethods("GET, POST, PUT, DELETE, OPTIONS")
	corscfg.AddAllowHeaders("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	corscfg.ExposeHeaders = []string{"Content-Length"}

	if !cfg.Server.Cors.AllowAll {
		corscfg.AllowAllOrigins = false
		corscfg.AllowOrigins = cfg.Server.Cors.Origin
	}

	srv.apiEngine.Use(
		cors.New(corscfg),
		middleware.RequestIDMiddleware(),
		middleware.TimeoutMiddleware(cfg.Server.WriteTimeout),
	)

	srv.apiServer = &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.Server.Port),
		Handler:      srv.apiEngine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infof("Start to rest api server :%v", srv.conf.Server.Auth.JWT.Key)
			return srv.Start()
		},
		OnStop: func(ctx context.Context) error {
			logger.Infof("Stopped rest api server")
			err := db.Close()
			if err != nil {
				logger.Warning(err)
			}
			return srv.Stop(ctx)
		},
	})
	return &srv, nil
}

func (srv *Server) Start() error {
	if !atomic.CompareAndSwapInt32(&srv.running, 0, 1) {
		return errors.New("server already started")
	}
	go func() {
		err := srv.apiServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.NewLogger().Warnf("failed to close http server", "err", err)
		}
	}()
	return nil
}

func (srv *Server) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&srv.running, 1, 0) {
		return errors.New("server already stopped")
	}

	var result error
	if err := srv.apiServer.Shutdown(ctx); err != nil {
		result = multierr.Append(result, fmt.Errorf("shutdown server. err: %v", err))
	}

	return result
}

func (srv *Server) RouteAPI() error {
	// Route common apis
	//srv.apiEngine.GET("version", func(gctx *gin.Context) {
	//	gctx.JSON(http.StatusOK, version.Get())
	//})
	//srv.apiEngine.GET("version", func(gctx *gin.Context) {
	//	gctx.JSON(http.StatusOK, "909090")
	//})
	// Route v1
	srv.apiEngine.RedirectTrailingSlash = false
	srv.apiEngine.RedirectFixedPath = false
	v1 := srv.apiEngine.Group("/api/v1")

	anonymousGroup := v1.Group("/")
	anonymousGroup.POST("/login", srv.authController.JWTMiddleware.LoginHandler)

	//anonymousGroup.POST("signup", handler.Wrap(srv.userController.HandleSignUp))
	//
	authGroup := v1.Group("/")
	authGroup.Use(srv.authController.AuthMiddleware())
	//
	courseGroup := authGroup.Group("course")
	////userGroup.POST("refresh-token", srv.authController.JWTMiddleware.RefreshHandler)
	courseGroup.GET("", handler.Wrap(srv.courseController.GetCourses))
	courseGroup.POST("", handler.Wrap(srv.courseController.AddCourse))
	courseGroup.PUT("/:id", handler.Wrap(srv.courseController.UpdateCourse))
	courseGroup.DELETE("/:id", handler.Wrap(srv.courseController.DeleteCourse))

	return nil
}
