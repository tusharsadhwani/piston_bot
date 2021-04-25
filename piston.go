package piston_bot

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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

type RunResponse struct {
	Result   string
	Language string
	Code     string
	Stdin    string
	Output   string
}

var stdinRegex = regexp.MustCompile(`\s\/stdin\b`)

func RunCode(update *tgbot.Update, text string) RunResponse {
	var lang, code string
	for index, char := range text {
		if char == ' ' || char == '\n' {
			lang, code = text[:index], text[index+1:]
			break
		}
	}
	if code == "" {
		return RunResponse{
			Result: ResultBadQuery,
			Output: "Bad Query",
		}
	}

	var stdin string
	stdinLoc := stdinRegex.FindStringIndex(code)
	if stdinLoc != nil {
		start, end := stdinLoc[0], stdinLoc[1]
		if end+1 < len(code) {
			code, stdin = code[:start], code[end+1:]
		}
	}

	jsonBody, err := json.Marshal(map[string]string{
		"language": lang,
		"version":  "",
		"files":    code,
		"stdin":    stdin,
	})
	if err != nil {
		log.Println(err)
		return RunResponse{
			Result: ResultUnknown,
		}
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://emkc.org/api/v2/piston/execute",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		log.Println(err)
		return RunResponse{
			Result: ResultUnknown,
		}
	}
	req.Header = http.Header{
		"Authorization": authHeader,
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if resp.Body != nil {
			body, err := ioutil.ReadAll(resp.Body)
			log.Println(err)
			log.Printf("%s\n", body)
		}
		log.Println(err)
		return RunResponse{
			Result: ResultUnknown,
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		log.Printf("%s\n", body)
		return RunResponse{
			Result: ResultUnknown,
		}
	}
	if resp.StatusCode != 200 {
		var errorStruct struct{ Message string }
		json.Unmarshal(body, &errorStruct)
		if errorStruct.Message == "" {
			log.Println(err)
			log.Printf("%s\n", body)
			return RunResponse{
				Result: ResultUnknown,
			}
		}

		return RunResponse{
			Result:   ResultError,
			Language: lang,
			Code:     code,
			Stdin:    stdin,
			Output:   errorStruct.Message,
		}
	}

	var data struct{ Run struct{ Output string } }
	json.Unmarshal(body, &data)
	return RunResponse{
		Result:   ResultSuccess,
		Language: lang,
		Code:     code,
		Stdin:    stdin,
		Output:   data.Run.Output,
	}
}
