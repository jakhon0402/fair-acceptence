package cmd

import (
	"fajr-acceptance/internal/config"
	"fajr-acceptance/internal/controller"
	"fajr-acceptance/internal/database"
	"fajr-acceptance/internal/server"
	"fajr-acceptance/internal/telegrambot"
	"fajr-acceptance/pkg/logger"
	"go.uber.org/fx"
)

func RunApplication() {
	cnf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log := logger.NewLogger()
	fx.New(
		fx.Supply(&cnf),
		fx.Supply(log),
		fx.Provide(
			database.NewMongoDb,
			controller.NewAuthController,
			controller.NewCourseController,
			telegrambot.NewTelegramBot,
			server.NewServer,
		),

		fx.Invoke(func(srv *server.Server, tb *telegrambot.TelegramBot) error {
			//logrus.Info(tb.BotToken)
			go tb.StartTelegramBot()
			return srv.RouteAPI()
		}),
	).Run()
}
