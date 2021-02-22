package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/getsentry/sentry-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

//Js -
type Js map[string]interface{}

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var client = &http.Client{}

func initMenu() {
	defer sentry.Recover()
	payload, _ := ioutil.ReadFile(`getstarted.json`)
	request, _ := http.NewRequest(`POST`, `https://graph.facebook.com/v9.0/me/messenger_profile?access_token=`+PageAccessToken,
		bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	payload, _ = ioutil.ReadAll(response.Body)
	if gjson.Get(string(payload), `error.message`).Exists() {
		panic(errors.New(string(payload)))
	}
}

func initPersistentMenu(psid string) {
	defer sentry.Recover()
	payload, _ := ioutil.ReadFile(`persistentmenu.json`)
	obj := Js{
		"psid":            psid,
		"persistent_menu": string(payload),
	}
	payload, _ = json.Marshal(obj)
	request, _ := http.NewRequest(`POST`, `https://graph.facebook.com/v9.0/me/custom_user_settings?access_token=`+PageAccessToken,
		bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	payload, _ = ioutil.ReadAll(response.Body)
	if gjson.Get(string(payload), `error.message`).Exists() {
		panic(errors.New(string(payload)))
	}
}

func sendRawMessage(psid string, obj interface{}) {
	jsonobj := Js{
		`recipient`: Js{
			`id`: psid,
		},
		`message`: obj,
	}
	payload, _ := json.Marshal(jsonobj)
	request, _ := http.NewRequest(`POST`, `https://graph.facebook.com/v9.0/me/messages?access_token=`+PageAccessToken,
		bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	payload, _ = ioutil.ReadAll(response.Body)
	if gjson.Get(string(payload), `error.message`).Exists() {
		panic(errors.New(string(payload)))
	}
}

func sendPostback(psid string, postback Postback) {
	go sendRawMessage(psid, Js{
		`attachment`: Js{
			`type`:    `template`,
			`payload`: postback,
		},
	})
}

func sendPostbackOrText(psid string, obj gjson.Result) {
	if obj.Get(`buttons`).Exists() {
		var postback Postback
		json.Unmarshal([]byte(obj.Raw), &postback)
		sendPostback(psid, postback)
	} else {
		sendText(psid, obj.Get(`text`).String())
	}
}

func sendText(psid string, texts ...interface{}) {
	fulltext := ``
	for _, text := range texts {
		fulltext += text.(string) + "\n"
	}
	go sendRawMessage(psid, Js{
		`text`: fulltext,
	})
}

func sendTextSync(psid string, texts ...interface{}) {
	fulltext := ``
	for _, text := range texts {
		fulltext += text.(string) + "\n"
	}
	sendRawMessage(psid, Js{
		`text`: fulltext,
	})
}

func sendAttachmentURL(psid string, attachmentType string, url string) {
	go sendRawMessage(psid, Js{
		`attachment`: Js{
			`type`: attachmentType,
			`payload`: Js{
				`url`:         url,
				`is_reusable`: true,
			},
		},
	})
}

func getGistFile(id string, filename string) gjson.Result {
	request, _ := http.NewRequest(`GET`, `https://api.github.com/gists/`+id, nil)
	request.Header.Add(`Accept`, `application/vnd.github.v3+json`)
	request.Header.Add(`Authorization`, `Basic `+
		base64.StdEncoding.EncodeToString([]byte(BasicAuth)))
	response, err := client.Do(request)

	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	payload, _ := ioutil.ReadAll(response.Body)
	if gjson.GetBytes(payload, `message`).Exists() {
		panic(errors.New(string(payload)))
	} else {
		result := gjson.GetBytes(payload, `files.`+filename+`.content`)
		if !result.Exists() {
			panic(`Gist file not found`)
		}
		return result
	}
}

func setGistFile(id string, filename string, obj interface{}) error {
	content, _ := json.Marshal(obj)
	payload, _ := json.Marshal(Js{
		`files`: Js{
			filename: Js{
				`content`: string(content),
			},
		},
	})
	request, _ := http.NewRequest(`PATCH`, `https://api.github.com/gists/`+id, bytes.NewBuffer(payload))
	request.Header.Add(`Accept`, `application/vnd.github.v3+json`)
	request.Header.Add(`Authorization`, `Basic `+
		base64.StdEncoding.EncodeToString([]byte(BasicAuth)))
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	payload, _ = ioutil.ReadAll(response.Body)
	if gjson.GetBytes(payload, `message`).Exists() {
		panic(errors.New(string(payload)))
	}
	return nil
}
