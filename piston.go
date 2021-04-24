package piston_bot

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	ResultSuccess  = "success"
	ResultError    = "error"
	ResultBadQuery = "badquery"
	ResultUnknown  = "unknown"
)

var authHeader []string

func Init() {
	authHeader = []string{os.Getenv("AUTH")}
}

func GetLanguages() ([]string, error) {
	resp, err := http.Get("https://emkc.org/api/v2/piston/runtimes")
	if err != nil {
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			log.Println(err)
			log.Printf("%s\n", body)
		}
		log.Println(err)
		return nil, err
	}
	var languagesMap []struct {
		Language string
		Version  string
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(body)
		log.Println(err)
		return nil, err
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
	return languages, nil
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
		log.Println(err)
		return
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://emkc.org/api/v2/piston/execute",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		result = ResultUnknown
		log.Println(err)
		return
	}
	req.Header = http.Header{
		"Authorization": authHeader,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result = ResultUnknown
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			log.Println(err)
			log.Printf("%s\n", body)
		}
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result = ResultUnknown
		log.Println(err)
		log.Printf("%s\n", body)
		return
	}
	if resp.StatusCode != 200 {
		var errorStruct struct{ Message string }
		json.Unmarshal(body, &errorStruct)
		if errorStruct.Message == "" {
			result = ResultUnknown

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
