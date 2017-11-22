package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/QtRoS/acadebot2/shared"
	"github.com/QtRoS/acadebot2/shared/logu"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	envAPIKey        = "ENV_API_KEY"
	perSrcLimit      = 12
	noCoursesFound   = "Sorry, no similar course found."
	dummyPlaceholder = "Just a moment..."
	noContextFound   = "Sorry, can't navigate through results. Try to search again!"
	greeting         = `Hello, %s!
I can help you with finding online courses (MOOCs).
Type course name or keyword and I will find something for you! 
(Works on Raspberry Pi 3)
	https://storebot.me/bot/acade_bot
	https://github.com/QtRoS/acadebot2`
)

var bot *tgbotapi.BotAPI

func init() {
	token := os.Getenv(envAPIKey)

	logu.Trace.Print("Token: ", token)

	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		logu.Error.Panic(err)
	}
	bot.Debug = false
	logu.Info.Printf("Authorized on account %s", bot.Self.UserName)
}

func main() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.InlineQuery != nil {
			handleInlineQuery(update.InlineQuery)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(update.CallbackQuery)
		} else if update.Message != nil {
			if update.Message.IsCommand() {
				handleCommand(update.Message)
			} else {
				handleMessage(update.Message)
			}
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	logu.Info.Printf("Message [%s] %s", message.From.UserName, message.Text)

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, dummyPlaceholder))

	query := strings.TrimSpace(message.Text)
	courses := getCourses(query)
	if courses == nil || len(courses) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, noCoursesFound)
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	} else {
		context := userContext{Query: query, Position: 0, Count: len(courses)}
		saveContext(message.Chat.ID, &context)

		courseInfo := courses[context.Position]

		msg := tgbotapi.NewMessage(message.Chat.ID, courseInfo.String())
		msg.ReplyToMessageID = message.MessageID
		msg.ReplyMarkup = createKeyboard(&context)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func handleCommand(message *tgbotapi.Message) {

	var answer string
	switch command := message.Command(); command {
	case "start":
		answer = fmt.Sprintf(greeting, message.From.UserName)
	default:
		answer = fmt.Sprintf("Unknown command: %s", command)
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, answer))
}

func handleInlineQuery(inlineQuery *tgbotapi.InlineQuery) {
	logu.Info.Printf("Inline [%s] %s", inlineQuery.From.UserName, inlineQuery.Query)
	courses := getCourses(inlineQuery.Query)
	if courses == nil || len(courses) == 0 {
		return
	}

	var articles = make([]interface{}, len(courses))
	for i, c := range courses {
		article := courseInfoToInlineQueryResult(c)
		articles[i] = article
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: inlineQuery.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       articles,
	}

	logu.Trace.Println("Articles:", len(articles))
	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		logu.Error.Println(err)
	}
}

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	// Dummy answer to stop spinners in UI.
	bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: callbackQuery.ID})

	// Check if there context for that user.
	context := restoreContext(callbackQuery.Message.Chat.ID)
	if context == nil {
		bot.Send(tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, noContextFound))
		return
	}

	// Calculate delta.
	delta, _ := strconv.Atoi(callbackQuery.Data)
	context.Position = shared.Max(context.Position+delta, 0)
	context.Position = shared.Min(context.Position, context.Count-1)
	saveContext(callbackQuery.Message.Chat.ID, context)

	// Get last results.
	courses := getCourses(context.Query)
	if len(courses) == 0 {
		logu.Warning.Println("getCourses returned no courses")
		return
	}
	courseInfo := courses[shared.Min(len(courses)-1, context.Position)]

	// Answer in TG.
	msg := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID, courseInfo.String())
	keyboard := createKeyboard(context)
	msg.BaseEdit.ReplyMarkup = &keyboard
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// ---------------------------------------- Helpers ==========================$

func getCourses(query string) []shared.CourseInfo {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return nil
	}

	jsonStr := RudraSearch(query, perSrcLimit)
	//logu.Trace.Println(jsonStr)

	var courses []shared.CourseInfo
	if err := json.Unmarshal([]byte(jsonStr), &courses); err != nil {
		logu.Error.Println("Bad JSON:", err)
		return nil
	}

	// for _, c := range courses {
	// 	logu.Trace.Println(c.Link)
	// }

	return courses
}

func courseInfoToInlineQueryResult(c shared.CourseInfo) tgbotapi.InlineQueryResultArticle {
	id := fmt.Sprintf("%x", md5.Sum([]byte(c.Link)))
	article := tgbotapi.NewInlineQueryResultArticle(id, c.Name, c.String())
	article.URL = c.Link
	article.ThumbURL = c.Art
	return article
}

func createKeyboard(context *userContext) tgbotapi.InlineKeyboardMarkup {
	status := fmt.Sprintf("%d of %d", context.Position+1, context.Count)
	btbl := tgbotapi.NewInlineKeyboardButtonData(status, "0")

	bm := tgbotapi.NewInlineKeyboardButtonData("◀ Previous", "-1")
	bp := tgbotapi.NewInlineKeyboardButtonData("Next   ▶", "+1")
	bfm := tgbotapi.NewInlineKeyboardButtonData("⏪", "-5")
	bfp := tgbotapi.NewInlineKeyboardButtonData("⏩", "+5")

	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(bm, bp),
		tgbotapi.NewInlineKeyboardRow(bfm, btbl, bfp))
}
