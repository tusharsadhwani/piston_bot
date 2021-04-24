package main

import (
	"fmt"
	"html"
	"log"
	"os"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	uuid "github.com/google/uuid"
	piston "github.com/tusharsadhwani/piston_bot"
)

var USAGE_MSG = `
<b>Usage:</b>
<pre>/run [language]
[your code]
...
</pre>
type /langs for list of supported languages.
`

var INLINE_USAGE_MSG = `
<b>Inline usage:</b>
<pre>@iruncode_bot [language]
[your code]
...
</pre>
`

var OUTPUT_MSG = `
<b>Language:</b>
<pre>%s</pre>

<b>Code:</b>
<pre>%s</pre>

<b>Output:</b>
<pre>%s</pre>
`

var ERROR_MSG = `
<b>Language:</b>
<pre>%s</pre>

<b>Code:</b>
<pre>%s</pre>

<b>Error:</b>
<pre>%s</pre>
`

var ERROR_STRING = "Some error occured, try again later."

func main() {
	piston.Init()

	bot, err := tgbot.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(err)
	}

	for update := range updates {
		if update.InlineQuery != nil {
			if update.InlineQuery.Query != "" {
				response := piston.RunCode(&update, update.InlineQuery.Query)
				var formattedText string
				switch response.Result {
				case piston.ResultBadQuery:
					formattedText = INLINE_USAGE_MSG
				case piston.ResultUnknown:
					formattedText = ERROR_STRING
				case piston.ResultError:
					formattedText = fmt.Sprintf(
						ERROR_MSG,
						html.EscapeString(response.Language),
						html.EscapeString(response.Code),
						html.EscapeString(response.Output),
					)
				case piston.ResultSuccess:
					formattedText = fmt.Sprintf(
						OUTPUT_MSG,
						html.EscapeString(response.Language),
						html.EscapeString(response.Code),
						html.EscapeString(response.Output),
					)
				}
				bot.AnswerInlineQuery(tgbot.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					Results: []interface{}{
						tgbot.InlineQueryResultArticle{
							Type:        "article",
							ID:          uuid.NewString(),
							Title:       "Output",
							Description: response.Output,
							InputMessageContent: tgbot.InputTextMessageContent{
								Text:      formattedText,
								ParseMode: "html",
							},
						},
					},
				})
			}
		}

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			msg := tgbot.NewMessage(update.Message.Chat.ID, "")
			msg.ParseMode = "html"
			msg.ReplyToMessageID = update.Message.MessageID

			switch update.Message.Command() {
			case "help":
				msg.Text = USAGE_MSG

			case "run":
				response := piston.RunCode(&update, update.Message.CommandArguments())
				switch response.Result {
				case piston.ResultBadQuery:
					msg.Text = USAGE_MSG
				case piston.ResultUnknown:
					msg.Text = ERROR_STRING
				case piston.ResultError:
					msg.Text = fmt.Sprintf(
						ERROR_MSG,
						html.EscapeString(response.Language),
						html.EscapeString(response.Code),
						html.EscapeString(response.Output),
					)
				case piston.ResultSuccess:
					msg.Text = fmt.Sprintf(
						OUTPUT_MSG,
						html.EscapeString(response.Language),
						html.EscapeString(response.Code),
						html.EscapeString(response.Output),
					)
				}

			case "langs":
				languages, err := piston.GetLanguages()
				if err != nil {
					msg.Text = ERROR_STRING
					break
				}
				textLines := make([]string, len(languages)+1)
				textLines = append(textLines, "<b>Supported languages:</b>")
				for _, lang := range languages {
					textLines = append(textLines, fmt.Sprintf("<pre>%s</pre>", html.EscapeString(lang)))
				}
				msg.Text = strings.Join(textLines, "\n")
			}
			bot.Send(msg)
		}
	}
}
