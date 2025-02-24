package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"sarasa/libs/retryHandling"

	"sarasa/libs/signals"

	"sarasa/libs/configHandling"

	"sarasa/schemas"

	"sarasa/libs/postgres"

	"github.com/google/uuid"
	"github.com/streadway/amqp"

	"sarasa/libs/influxdb"
	"sarasa/libs/rabbitMQ"

	"sarasa/libs/errorHandling"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var configuration schemas.Config

var postgresSingleton postgres.Client
var rabbitMQSingleton rabbitMQ.Client
var influxSingleton influxdb.Client

var availableZones map[string]int
var availableProviders []schemas.Provider

type botClient struct {
	bot *tgbotapi.BotAPI
}

func init() {
	err := configHandling.LoadConfig(&configuration, "telegram")
	errorHandling.FailOnError(err, "Could not get configuration")

	/**
	 * RabbitMQ Singleton
	 */
	err = rabbitMQSingleton.Init(configuration.RabbitMQ)
	errorHandling.FailOnError(err, "Could not initialize rabbitMQSingleton")

	/**
	 * Postgres Singleton
	 */
	errorHandling.FailOnError(
		postgresSingleton.Init(configuration.Postgres), "Could not initialize PostgresSingleton")

	availableZones, err = postgresSingleton.GetZones()
	errorHandling.FailOnError(err, "Could not load zones")

	/**
	 * InfluxDB Singleton
	 */
	if err := influxSingleton.Init(configuration.Influx); err != nil {
		log.Printf("Warn - Could not initialize InfluxSingleton, error: %s", err)
	}

	/**
	 * Signal handling
	 */
	signals.SignalHandler(rabbitMQSingleton, postgresSingleton, influxSingleton)
}

func main() {
	var err error

	botClient := botClient{}

	botClient.bot, err = tgbotapi.NewBotAPI(configuration.Telegram.Token)
	errorHandling.FailOnError(err, "Failed to create bot instance")

	botClient.bot.Debug = true

	log.Printf("Authorized on account %s", botClient.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := botClient.bot.GetUpdatesChan(u)
	errorHandling.FailOnError(err, "Failed to create update channel")

	for update := range updates {
		go botClient.handleUpdate(update)
	}
}

func (botClient botClient) handleUpdate(update tgbotapi.Update) {
	var updateType string

	if update.Message != nil || update.CallbackQuery != nil { // ignore any non-Message Updates
		startTime := time.Now()

		log.Printf("Received update: %#v", update)

		if update.Message != nil {
			log.Printf("Username: %s - Text: %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				updateType = "command"
				botClient.handleCommand(update.Message)
			} else {
				updateType = "message"
				botClient.handleMessage(update.Message)
			}
		}

		if update.CallbackQuery != nil {
			updateType = "callback_query"
			botClient.handleCallbackQuery(update.CallbackQuery)
		}

		influxTags := map[string]string{"run_uuid": uuid.New().String(), "update_type": updateType}
		influxFields := map[string]interface{}{}

		influxFields = map[string]interface{}{
			"elapsed": time.Since(startTime).Milliseconds(),
			"success": true,
		}

		go errorHandling.LogOnError(
			influxSingleton.Send("telegram_process", influxTags, influxFields),
			"Could not write to InfluxDB")
	}
}

func (botClient botClient) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	var err error

	if len(availableProviders) == 0 {
		availableProviders, err = postgresSingleton.GetProviders()
		errorHandling.FailOnError(err, "Could not load providers")
	}

	// @TODO re think this...
	if strings.Contains(callbackQuery.Message.Text, "Providers from ") {
		for _, provider := range availableProviders {
			selectedProviderID, err := strconv.Atoi(callbackQuery.Data) // Data is the ProviderID
			errorHandling.FailOnError(err, "Error converting selected provider ID to int")

			if provider.ID == selectedProviderID {
				msg := tgbotapi.NewMessage(
					callbackQuery.Message.Chat.ID,
					fmt.Sprintf("Providers name: %s - Whatsapp: %s",
						provider.Name,
						fmt.Sprintf("https://wa.me/549%s", provider.Phone),
					),
				)

				_, err = botClient.bot.Send(msg)
				errorHandling.LogOnError(err, "[handleCallbackQuery] Error sending message to Telegram")

				mediaFiles := make([]interface{}, 0)
				for index, pic := range provider.Pics {
					if index == 10 { // @TODO There should be a best way with queries
						break
					}

					mediaFiles = append(mediaFiles, tgbotapi.NewInputMediaPhoto(pic))
				}

				cfg := tgbotapi.NewMediaGroup(callbackQuery.Message.Chat.ID, mediaFiles)

				_, err = botClient.bot.Send(cfg)
				errorHandling.LogOnError(err, "[handleCallbackQuery 1] Error sending message to Telegram")

				break
			}
		}
	} else {
		var inlineKeyboardRow []tgbotapi.InlineKeyboardButton
		replay := tgbotapi.NewInlineKeyboardMarkup()

		maxButtons := 3 // @TODO Hardcoded logic
		for _, provider := range availableProviders {
			if provider.Place != callbackQuery.Data {
				continue
			}

			inlineKeyboardRow = append(
				inlineKeyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(provider.Name, strconv.Itoa(provider.ID)),
			)
			maxButtons--

			if maxButtons == 0 {
				replay.InlineKeyboard = append(replay.InlineKeyboard, inlineKeyboardRow)
				inlineKeyboardRow = tgbotapi.NewInlineKeyboardRow()

				maxButtons = 3
			}
		}

		replay.InlineKeyboard = append(replay.InlineKeyboard, inlineKeyboardRow)

		msg := tgbotapi.NewMessage(
			callbackQuery.Message.Chat.ID,
			fmt.Sprintf("No providers from %s", callbackQuery.Data),
		)

		if len(inlineKeyboardRow) != 0 {
			msg.ReplyMarkup = replay
			msg.Text = fmt.Sprintf("Providers from %s", callbackQuery.Data)
		}

		_, err = botClient.bot.Send(msg)
		errorHandling.LogOnError(err, "[handleCallbackQuery 2] Error sending message to Telegram")
	}
}

func (botClient botClient) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "refresh":
		botClient.commandRefresh(message)
	case "get_by_zone":
		botClient.commandGetByZone(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Invalid command")
		msg.ReplyToMessageID = message.MessageID

		_, err := botClient.bot.Send(msg)
		errorHandling.LogOnError(err, "[handleCommand] Error sending message to Telegram")
	}
}

func (botClient botClient) handleMessage(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
	msg.ReplyToMessageID = message.MessageID

	_, err := botClient.bot.Send(msg)
	errorHandling.LogOnError(err, "[handleMessage] Error sending message to Telegram")
}

func (botClient botClient) commandRefresh(message *tgbotapi.Message) {
	var err error

	errorHandling.FailOnError(
		retryHandling.Try(
			func() error {
				return rabbitMQSingleton.Channel.ExchangeDeclare(
					"refresh", "fanout", true,
					false, false, false, nil)
			},
			[]error{amqp.ErrClosed},
			func() error { return rabbitMQSingleton.Restart() },
			2, 5000*time.Millisecond,
		),
		"Failed to try ExchangeDeclare",
	)

	errorHandling.FailOnError(
		rabbitMQSingleton.Channel.Publish(
			"refresh", "", false, false, amqp.Publishing{Body: []byte{}}),
		"Failed to publish refresh message",
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Processing refresh...")
	msg.ReplyToMessageID = message.MessageID

	availableProviders = make([]schemas.Provider, 0)

	_, err = botClient.bot.Send(msg)
	errorHandling.LogOnError(err, "[commandRefresh] Error sending message to Telegram")
}

func (botClient botClient) commandGetByZone(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Zones with providers")
	msg.ReplyToMessageID = message.MessageID

	var inlineKeyboardRow []tgbotapi.InlineKeyboardButton
	replay := tgbotapi.NewInlineKeyboardMarkup()

	maxButtons := 3
	for zone := range availableZones {
		inlineKeyboardRow = append(inlineKeyboardRow, tgbotapi.NewInlineKeyboardButtonData(zone, zone))
		maxButtons--

		if maxButtons == 0 {
			replay.InlineKeyboard = append(replay.InlineKeyboard, inlineKeyboardRow)
			inlineKeyboardRow = tgbotapi.NewInlineKeyboardRow()

			maxButtons = 3
		}
	}

	if len(inlineKeyboardRow) != 0 {
		replay.InlineKeyboard = append(replay.InlineKeyboard, inlineKeyboardRow)
	}

	msg.ReplyMarkup = replay

	_, err := botClient.bot.Send(msg)
	errorHandling.LogOnError(err, "[commandGetByZone] Error sending message to Telegram")
}
