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

var OUTPUT_MSG = `
<b>Code:</b>
<pre>%s</pre>

<b>Output:</b>
<pre>%s</pre>
`

var ERROR_MSG = `
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

	// bot.Debug = true

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
				result, code, text := piston.RunCode(&update, update.InlineQuery.Query)
				var formattedText string
				switch result {
				case piston.ResultBadQuery:
					formattedText = USAGE_MSG
				case piston.ResultUnknown:
					formattedText = ERROR_STRING
				case piston.ResultError:
					formattedText = fmt.Sprintf(ERROR_MSG, html.EscapeString(code), html.EscapeString(text))
				case piston.ResultSuccess:
					formattedText = fmt.Sprintf(OUTPUT_MSG, html.EscapeString(code), html.EscapeString(text))
				}
				bot.AnswerInlineQuery(tgbot.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					Results: []interface{}{
						tgbot.InlineQueryResultArticle{
							Type:        "article",
							ID:          uuid.NewString(),
							Title:       "Output",
							Description: text,
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
				result, code, text := piston.RunCode(&update, update.Message.CommandArguments())
				switch result {
				case piston.ResultBadQuery:
					msg.Text = USAGE_MSG
				case piston.ResultUnknown:
					msg.Text = ERROR_STRING
				case piston.ResultError:
					msg.Text = fmt.Sprintf(ERROR_MSG, html.EscapeString(code), html.EscapeString(text))
				case piston.ResultSuccess:
					msg.Text = fmt.Sprintf(OUTPUT_MSG, html.EscapeString(code), html.EscapeString(text))
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
