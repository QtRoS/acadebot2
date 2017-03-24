package main

import (
	"../shared"
	"../shared/logu"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"strconv"
	"strings"
)

const (
	FileName       = "api.key"
	PerSrcLimit    = 10
	NoCoursesFound = "Sorry, no similar course found"
)

var token string
var bot *tgbotapi.BotAPI

func init() {
	content, err := ioutil.ReadFile(FileName)
	if err != nil {
		logu.Error.Printf("Can't read token:", err)
	}
	token = string(content)
	logu.Trace.Print("Token: ", token)

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
		} else if update.Message != nil {
			handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(update.CallbackQuery)
		}
	}
}

func handleMessage(message *tgbotapi.Message) {
	logu.Info.Printf("[%s] %s", message.From.UserName, message.Text)
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
		msg.ReplyMarkup = createKeyboard()
		bot.Send(msg)
	}
}

func handleInlineQuery(inlineQuery *tgbotapi.InlineQuery) {
	logu.Info.Printf("[%s] %s", inlineQuery.From.UserName, inlineQuery.Query)
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

	delta, _ := strconv.Atoi(callbackQuery.Data)

	context := RestoreContext(callbackQuery.Message.Chat.ID)
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
	keyboard := createKeyboard()
	msg.BaseEdit.ReplyMarkup = &keyboard
	bot.Send(msg)
}

func getCourses(query string) []shared.CourseInfo {
	query = strings.TrimSpace(query)
	if len(query) == 0 {
		return nil
	}

	jsonStr := Search(query, PerSrcLimit)
	logu.Trace.Println(jsonStr)
	courses := make([]shared.CourseInfo, 0, PerSrcLimit)

	if err := json.Unmarshal([]byte(jsonStr), &courses); err != nil {
		logu.Error.Println("Bad JSON:", err)
		return nil
	}

	return courses
}

func courseInfoToInlineQueryResult(c shared.CourseInfo) tgbotapi.InlineQueryResultArticle {
	id := fmt.Sprintf("%x", md5.Sum([]byte(c.Link)))
	article := tgbotapi.NewInlineQueryResultArticleMarkdown(id, c.Name, c.String())
	article.URL = c.Link
	article.ThumbURL = c.Art
	return article
}

func createKeyboard() tgbotapi.InlineKeyboardMarkup {

	bm := tgbotapi.NewInlineKeyboardButtonData("<", "-1")
	bp := tgbotapi.NewInlineKeyboardButtonData(">", "+1")
	bfm := tgbotapi.NewInlineKeyboardButtonData("<<", "-5")
	bfp := tgbotapi.NewInlineKeyboardButtonData(">>", "+5")

	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(bm, bp),
		tgbotapi.NewInlineKeyboardRow(bfm, bfp))
}

// def create_nav_keyboard():
//     markup = types.InlineKeyboardMarkup(row_width=2)
//     bm = types.InlineKeyboardButton('Back', callback_data=str(-1))
//     bp = types.InlineKeyboardButton('Forward', callback_data=str(1))
//     bmm = types.InlineKeyboardButton('Fast back', callback_data=str(-5))
//     bpp = types.InlineKeyboardButton('Fast forward', callback_data=str(5))
//     markup.add(bm, bp, bmm, bpp)
//     return markup
