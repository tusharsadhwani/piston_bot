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
/stdin [input text] (optional)
...</pre>

type /langs for list of supported languages.
`

var INLINE_USAGE_MSG = `
<b>Inline usage:</b>
<pre>@iruncode_bot [language]
[your code]
...
/stdin [input text] (optional)
...</pre>
`

var ERROR_STRING = `
Some error occured, try again later.
If the error persists, report it to the admins in the bot's bio.
`

func main() {
	piston.Init()

	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Println("Unable to read bot token. Make sure you export $TOKEN in the environment.")
		os.Exit(1)
	}

	bot, err := tgbot.NewBotAPI(token)
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
				var response piston.RunResponse
				var title string
				var message string
				request, err := piston.CreateRequest(update.InlineQuery.Query)
				if err != nil {
					title = "Bad Query"
					message = INLINE_USAGE_MSG
				} else {
					response = piston.RunCode(&update, request)
					title = response.Output
					message = formatPistonResponse(request, response)
				}

				bot.AnswerInlineQuery(tgbot.InlineConfig{
					InlineQueryID: update.InlineQuery.ID,
					Results: []interface{}{
						tgbot.InlineQueryResultArticle{
							Type:        "article",
							ID:          uuid.NewString(),
							Title:       "Output",
							Description: title,
							InputMessageContent: tgbot.InputTextMessageContent{
								Text:      message,
								ParseMode: "html",
							},
							ReplyMarkup: forkButton(request),
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
				request, err := piston.CreateRequest(update.Message.CommandArguments())
				if err != nil {
					msg.Text = USAGE_MSG
					break
				}

				response := piston.RunCode(&update, request)
				msg.Text = formatPistonResponse(request, response)
				msg.ReplyMarkup = forkButton(request)

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

var (
	BlockLanguage = "Language"
	BlockCode     = "Code"
	BlockStdin    = "Stdin"
	BlockOutput   = "Output"
	BlockError    = "Error"
)
var blockNames = []string{BlockLanguage, BlockCode, BlockStdin, BlockOutput, BlockError}

func buildOutput(blocks map[string]string) string {
	var formattedBlocks []string
	for _, blockName := range blockNames {
		blockText := blocks[blockName]
		if blockText != "" {
			formattedName := fmt.Sprintf("<b>%s:</b>", blockName)
			formattedText := fmt.Sprintf("<pre>%s</pre>", html.EscapeString(blockText))

			formattedBlock := formattedName + "\n" + formattedText
			formattedBlocks = append(formattedBlocks, formattedBlock)
		}
	}

	return strings.Join(formattedBlocks, "\n\n")
}

func formatPistonResponse(request piston.RunRequest, response piston.RunResponse) string {
	switch response.Result {
	case piston.ResultUnknown:
		return ERROR_STRING

	case piston.ResultError:
		return buildOutput(map[string]string{
			BlockLanguage: request.Language,
			BlockCode:     request.Code,
			BlockStdin:    request.Stdin,
			BlockError:    response.Output,
		})

	case piston.ResultSuccess:
		return buildOutput(map[string]string{
			BlockLanguage: request.Language,
			BlockCode:     request.Code,
			BlockStdin:    request.Stdin,
			BlockOutput:   response.Output,
		})
	}

	return ""
}

func forkButton(request piston.RunRequest) *tgbot.InlineKeyboardMarkup {
	forkText := request.Language + "\n" + request.Code
	inlineKeyboard := tgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbot.InlineKeyboardButton{
			{
				tgbot.InlineKeyboardButton{
					Text:                         "Fork",
					SwitchInlineQueryCurrentChat: &forkText,
				},
			},
		},
	}
	return &inlineKeyboard
}
