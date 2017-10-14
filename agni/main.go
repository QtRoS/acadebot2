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
	EnvApiKey        = "ENV_API_KEY"
	PerSrcLimit      = 10
	NoCoursesFound   = "Sorry, no similar course found."
	DummyPlaceholder = "Just a moment..."
	NoContextFound   = "Sorry, can't navigate through results. Try to search again!"
	Greeting         = `Hello, %s!
	I can help you with finding online courses (MOOCs).
	Type course name or keyword and I will find something for you! 
	https://storebot.me/bot/acade_bot`
)

var bot *tgbotapi.BotAPI

func init() {
	token := os.Getenv(EnvApiKey)

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

	bot.Send(tgbotapi.NewMessage(message.Chat.ID, DummyPlaceholder))

	query := strings.TrimSpace(message.Text)
	courses := getCourses(query)
	if courses == nil || len(courses) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, NoCoursesFound)
		msg.ReplyToMessageID = message.MessageID
		bot.Send(msg)
	} else {
		context := UserContext{Query: query, Position: 0, Count: len(courses)}
		SaveContext(message.Chat.ID, &context)

		courseInfo := courses[context.Position]

		msg := tgbotapi.NewMessage(message.Chat.ID, courseInfo.String())
		msg.ReplyToMessageID = message.MessageID
		msg.ReplyMarkup = createKeyboard(&context)
		bot.Send(msg)
	}
}

func handleCommand(message *tgbotapi.Message) {

	var answer string
	switch command := message.Command(); command {
	case "start":
		answer = fmt.Sprintf(Greeting, message.From.UserName)
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
	context := RestoreContext(callbackQuery.Message.Chat.ID)
	if context == nil {
		bot.Send(tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, NoContextFound))
		return
	}

	delta, _ := strconv.Atoi(callbackQuery.Data)
	context.Position = shared.Max(context.Position+delta, 0)
	context.Position = shared.Min(context.Position, context.Count-1)
	SaveContext(callbackQuery.Message.Chat.ID, context)

	courses := getCourses(context.Query)
	if len(courses) == 0 {
		logu.Warning.Println("getCourses returned no courses")
		return
	}
	courseInfo := courses[shared.Min(len(courses)-1, context.Position)]

	msg := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID, courseInfo.String())
	keyboard := createKeyboard(context)
	msg.BaseEdit.ReplyMarkup = &keyboard
	bot.Send(msg)
}

// ---------------------------------------- Helpers ==========================$

func getCourses(query string) []shared.CourseInfo {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return nil
	}

	jsonStr := Search(query, PerSrcLimit)
	//logu.Trace.Println(jsonStr)
	courses := make([]shared.CourseInfo, 0, PerSrcLimit)

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

func createKeyboard(context *UserContext) tgbotapi.InlineKeyboardMarkup {
	status := fmt.Sprintf("%d of %d", context.Position+1, context.Count)
	btbl := tgbotapi.NewInlineKeyboardButtonData(status, "0")

	bm := tgbotapi.NewInlineKeyboardButtonData("◀ Previous", "-1")
	bp := tgbotapi.NewInlineKeyboardButtonData("Next   ▶", "+1")
	bfm := tgbotapi.NewInlineKeyboardButtonData("⏪", "-5")
	bfp := tgbotapi.NewInlineKeyboardButtonData("⏩", "+5")

	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(bm, bp),
		tgbotapi.NewInlineKeyboardRow(bfm, btbl, bfp))
}
