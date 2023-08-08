package telegrambot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var StartKeyboardsNotRegistered = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(COURSES),
		tgbotapi.NewKeyboardButton(LOCATION),
		tgbotapi.NewKeyboardButton(CONTACT),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(REGISTER),
	),
)

var StartKeyboardsRegistered = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(COURSES),
		tgbotapi.NewKeyboardButton(LOCATION),
		tgbotapi.NewKeyboardButton(CONTACT),
	),
)

var PrevKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(PREVIOUS)))
