package piston_bot

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	ResultSuccess  = "success"
	ResultError    = "error"
	ResultBadQuery = "badquery"
	ResultUnknown  = "unknown"
)

func GetLanguages() (output string) {
	resp, err := http.Get("https://emkc.org/api/v2/piston/runtimes")
	if err != nil {
		output = "some error occured, try again later."
		log.Println(err)
		return
	}
	var languagesMap []struct {
		Language string
		Version  string
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		output = "some error occured, try again later."
		log.Println(err)
		return
	}
	json.Unmarshal(body, &languagesMap)
	languageSet := make(map[string]struct{})
	for _, obj := range languagesMap {
		languageSet[obj.Language] = struct{}{}
	}
	var languages []string
	for lang := range languageSet {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	output = strings.Join(languages, "\n")
	return
}

func RunCode(update *tgbot.Update, text string) (result string, source string, output string) {
	if len(text) == 0 {
		result = ResultBadQuery
		output = "Bad Query"
		return
	}

	var lang, code string
	for index, char := range text {
		if char == ' ' || char == '\n' {
			lang, code = text[:index], text[index+1:]
			break
		}
	}
	if code == "" {
		result = ResultBadQuery
		output = "Bad Query"
		return
	}

	jsonBody, err := json.Marshal(map[string]string{
		"language": lang,
		"version":  "",
		"files":    code,
	})
	if err != nil {
		result = ResultUnknown
		output = "some error occured, try again later."
		log.Println(err)
		return
	}

	resp, err := http.Post(
		"https://emkc.org/api/v2/piston/execute",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		result = ResultUnknown
		output = "some error occured, try again later."
		body, err := io.ReadAll(resp.Body)
		log.Println(err)
		log.Printf("%s\n", body)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result = ResultUnknown
		output = "some error occured, try again later."
		log.Println(err)
		log.Printf("%s\n", body)
		return
	}
	if resp.StatusCode != 200 {
		var errorStruct struct{ Message string }
		json.Unmarshal(body, &errorStruct)
		if errorStruct.Message == "" {
			result = ResultUnknown
			output = "some error occured, try again later."
			log.Println(err)
			log.Printf("%s\n", body)
			return
		}
		result = ResultError
		source = code
		output = errorStruct.Message
		return
	}
	var data struct{ Run struct{ Output string } }
	json.Unmarshal(body, &data)
	result = ResultSuccess
	source = code
	output = data.Run.Output
	return
}
