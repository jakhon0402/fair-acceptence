package telegrambot

import (
	"context"
	"fajr-acceptance/internal/config"
	"fajr-acceptance/internal/database"
	"fajr-acceptance/internal/models"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

const (
	StudentCollection = "students"
	CourseCollection  = "courses"
)

type TelegramBot struct {
	BotToken string
	DB       *database.MongoDBClient
}

func NewTelegramBot(cnf *config.Config, logger *logrus.Logger, db *database.MongoDBClient) (*TelegramBot, error) {
	telegramBot := TelegramBot{
		BotToken: cnf.Server.BotToken,
		DB:       db,
	}

	//lc.Append(fx.Hook{
	//	OnStart: func(ctx context.Context) error {
	//		return telegramBot.StartTelegramBot()
	//	},
	//	OnStop: func(ctx context.Context) error {
	//		return nil
	//	},
	//})

	return &telegramBot, nil
}

func (tb *TelegramBot) StartTelegramBot() {
	bot, err := tgbotapi.NewBotAPI(tb.BotToken)
	if err != nil {
		fmt.Println(tb.BotToken)
		//log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	coll := tb.DB.GetCollection(CourseCollection)

	usersData := make(map[int64]models.Student)

	for update := range updates {

		if update.Message != nil {

			if update.Message.Text == "" {
				continue
			} // ignore any non-Message updates

			logrus.Info(update.Message.Text)
			currentUser := tb.getCurrentUser(update.Message.Chat.ID)

			// Create a new MessageConfig. We don't have text yet,
			// so we leave it empty.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			if update.Message.Command() == START {
				msg.Text = StartText
				if currentUser.IsRegistered {
					msg.ReplyMarkup = StartKeyboardsRegistered
				} else {
					msg.ReplyMarkup = StartKeyboardsNotRegistered
				}
				tb.changeState(START, currentUser)
			}

			if update.Message.Text == PREVIOUS {
				msg.Text = PrevText
				if currentUser.IsRegistered {
					msg.ReplyMarkup = StartKeyboardsRegistered
				} else {
					msg.ReplyMarkup = StartKeyboardsNotRegistered
				}
				tb.changeState(START, currentUser)
				if _, err := bot.Send(msg); err != nil {
					log.Panic(err)
				}
				continue
			}

			if currentUser.State == START {
				switch update.Message.Text {
				case LOCATION:
					msg.Text = "Lokatsiyamiz"
					msg.ReplyMarkup = PrevKeyboard
				case CONTACT:
					msg.Text = "Tel raqamlarimiz"
					msg.ReplyMarkup = PrevKeyboard

				case COURSES:
					coursesText := ""
					msg.ReplyMarkup = PrevKeyboard
					coursesButton := []tgbotapi.KeyboardButton{}

					cursor, _ := coll.Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))

					var courses []models.Course
					cursor.All(context.Background(), &courses)

					for index, val := range courses {
						coursesButton = append(coursesButton, tgbotapi.NewKeyboardButton(val.Name))
						if index == len(courses)-1 {
							coursesText += fmt.Sprintf("%v.\n", val.Name)

						} else {
							coursesText += fmt.Sprintf("%v,\n", val.Name)

						}
					}
					msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
						tgbotapi.NewKeyboardButtonRow(
							coursesButton...,
						),
						tgbotapi.NewKeyboardButtonRow(
							tgbotapi.NewKeyboardButton(PREVIOUS),
						),
					)
					tb.changeState(COURSES, currentUser)
					msg.Text = fmt.Sprintf("💻 Bizning kurslarimiz \n\n%v\n\n✅ Kurslarimiz haqida batafsil ma'lumotni quyida menu orqali bilib olasiz!", coursesText)

				case REGISTER:
					//cursor, _ := coll.Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))
					//
					//var courses []models.Course
					//cursor.All(context.Background(), &courses)
					//
					//for _, course := range courses {
					//	if course.Name == update.Message.Text {
					if currentUser.IsRegistered {
						msg.Text = AlreadyRegisteredText
					} else {
						tb.changeState(ENTER_FIRST_NAME, currentUser)
						msg.Text = EnterFirstnameText
						msg.ReplyMarkup = PrevKeyboard
						usersData[update.Message.Chat.ID] = models.Student{}
					}

					//
					//	}
					//}

				}
			} else if currentUser.State == ENTER_FIRST_NAME {
				if len(update.Message.Text) < 3 {
					msg.Text = InvalidFirstnameText
				} else {
					for key, val := range usersData {
						if key == update.Message.Chat.ID {
							val.FirstName = update.Message.Text
							usersData[key] = val
							break
						}
					}
					tb.changeState(ENTER_LAST_NAME, currentUser)
					msg.Text = EnterLastnameText
				}

			} else if currentUser.State == ENTER_LAST_NAME {
				if len(update.Message.Text) < 3 {
					msg.Text = InvalidLastnameText
				} else {
					for key, val := range usersData {
						if key == update.Message.Chat.ID {
							val.LastName = update.Message.Text
							usersData[key] = val
							if val.FirstName != "" {
								tb.changeState(ENTER_PHONE_NUMBER, currentUser)
								//msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
								//	tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButtonContact("Telefon raqamingizni kiritish")))
								msg.Text = EnterPhoneNumberText
							} else {
								tb.changeState(ENTER_FIRST_NAME, currentUser)
							}
							break

						}
					}

				}
			} else if currentUser.State == ENTER_PHONE_NUMBER {
				logrus.Info(update.Message.Text)
				if len(update.Message.Text) < 3 {
					msg.Text = InvalidPhoneNumberText
				} else {
					for key, val := range usersData {
						if key == update.Message.Chat.ID {

							if val.FirstName == "" || val.LastName == "" {
								tb.changeState(ENTER_FIRST_NAME, currentUser)
								break
							}
							coll := tb.DB.GetCollection(StudentCollection)

							filter := bson.D{{"chatId", currentUser.ChatId}}
							currentUser.FirstName = val.FirstName
							currentUser.LastName = val.LastName
							currentUser.PhoneNumber = update.Message.Text
							currentUser.IsRegistered = true
							update := bson.M{
								"$set": currentUser,
							}

							coll.UpdateOne(context.Background(), filter, update)

							tb.changeState(START, currentUser)
							msg.Text = fmt.Sprintf("🎉 Tabriklaymiz %v, %v ro'yhatdan o'tdingiz!", val.FirstName, val.LastName)
							msg.ReplyMarkup = StartKeyboardsRegistered
							break
						}
					}

				}
			} else if currentUser.State == COURSES {
				cursor, _ := coll.Find(context.Background(), bson.M{}, options.Find().SetSort(bson.M{"createdAt": -1}))

				var courses []models.Course
				cursor.All(context.Background(), &courses)

				for _, course := range courses {
					if course.Name == update.Message.Text {
						msg.Text = course.Description
						msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
							tgbotapi.NewInlineKeyboardRow(
								tgbotapi.NewInlineKeyboardButtonData("Kursga yozilish", course.ID.Hex()),
							),
						)
					}
				}
			}

			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		} else if update.CallbackQuery != nil {

			currentUser := tb.getCurrentUser(update.CallbackQuery.Message.Chat.ID)

			if !currentUser.IsRegistered {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, PleaseRegisterText)
				if _, err := bot.Request(callback); err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, PleaseRegisterText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
				continue
			}

			courseID, _ := primitive.ObjectIDFromHex(update.CallbackQuery.Data)
			filter := bson.D{{"_id", courseID}}

			var courseById models.Course
			err := tb.DB.GetCollection(CourseCollection).FindOne(context.Background(), filter).Decode(&courseById)
			if err != nil {
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, CourseNotFoundText)
				if _, err := bot.Request(callback); err != nil {
					panic(err)
				}
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, CourseNotFoundText)
				if _, err := bot.Send(msg); err != nil {
					panic(err)
				}
				continue
			}

			coll := tb.DB.GetCollection(StudentCollection)

			filter1 := bson.D{{"chatId", currentUser.ChatId}}

			accepted := false

			for _, val := range currentUser.Courses {
				if val.ID == courseById.ID {
					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, AlreadyAcceptedText)
					if _, err := bot.Request(callback); err != nil {
						panic(err)
					}
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, AlreadyAcceptedText)
					if _, err := bot.Send(msg); err != nil {
						panic(err)
					}
					accepted = true
					break
				}
			}
			if accepted {
				continue
			}

			currentUser.Courses = append(currentUser.Courses, courseById)
			updatecha := bson.M{
				"$set": currentUser,
			}

			coll.UpdateOne(context.Background(), filter1, updatecha)

			// Respond to the callback query, telling Telegram to show the user
			// a message with the data received.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, AcceptedDoneText)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}

			// And finally, send a message containing the data received.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, AcceptedDoneText)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func (tb *TelegramBot) getCurrentUser(chatId int64) (currentUser models.Student) {
	coll := tb.DB.GetCollection(StudentCollection)

	filter := bson.D{{"chatId", chatId}}

	var studentById models.Student
	err := coll.FindOne(context.Background(), filter).Decode(&studentById)
	if err != nil {

		student := models.Student{
			State:        "START",
			ChatId:       chatId,
			CreatedAt:    time.Now(),
			IsRegistered: false,
		}

		res, _ := coll.InsertOne(context.TODO(), student)

		var insertedStudent models.Student

		coll.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&insertedStudent)

		currentUser = insertedStudent
	} else {
		currentUser = studentById
	}
	return
}

func (tb *TelegramBot) changeState(state string, studentByChatID models.Student) {
	coll := tb.DB.GetCollection(StudentCollection)
	filter := bson.D{{"_id", studentByChatID.ID}}

	studentByChatID.State = state

	update := bson.M{
		"$set": studentByChatID,
	}

	coll.UpdateOne(context.Background(), filter, update)
}
