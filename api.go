package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/ahmetb/go-linq"
	"github.com/getsentry/sentry-go"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

//Js -
type Js map[string]interface{}

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var client = &http.Client{}

func initMenu(configFile string) {
	defer sentry.Recover()
	payload, _ := ioutil.ReadFile(configFile)
	request, _ := http.NewRequest(`POST`, `https://graph.facebook.com/v9.0/me/messenger_profile?access_token=`+PageAccessToken,
		bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	payload, _ = ioutil.ReadAll(response.Body)
	if errmess := gjson.Get(string(payload), `error.message`); errmess.Exists() {
		panic(errors.New(errmess.String()))
	}
}
func sendRawMessages(psid string, objs ...interface{}) {
	jsonobj := Js{
		`recipient`: Js{
			`id`: psid,
		},
	}
	for _, obj := range objs {
		jsonobj[`message`] = obj
		payload, _ := json.Marshal(jsonobj)
		request, _ := http.NewRequest(`POST`, `https://graph.facebook.com/v9.0/me/messages?access_token=`+PageAccessToken,
			bytes.NewBuffer(payload))
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		payload, _ = ioutil.ReadAll(response.Body)
		if errmess := gjson.Get(string(payload), `error.message`); errmess.Exists() {
			panic(errors.New(errmess.String()))
		}
	}
}

func sendPostback(psid string, postback Postback) {
	sendRawMessages(psid, Js{
		`attachment`: Js{
			`type`:    `template`,
			`payload`: postback,
		},
	})
}

func sendQuestion(psid string, questions []gjson.Result, counter int) {
	if questions[counter].Get(`buttons`).Exists() {
		var postback Postback
		json.Unmarshal([]byte(questions[counter].Raw), &postback)
		sendPostback(psid, postback)
	} else {
		sendText(psid, questions[counter].Get(`text`).String())
	}
}

func sendText(psid string, texts ...interface{}) {
	var messages []interface{}
	linq.From(texts).Select(func(v interface{}) interface{} {
		return Js{
			`text`: v,
		}
	}).ToSlice(&messages)
	sendRawMessages(psid, messages...)
}

func sendAttachmentURL(psid string, attachmentType string, url string) {
	sendRawMessages(psid, Js{
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
	if errormessage := gjson.GetBytes(payload, `message`); errormessage.Exists() {
		panic(errormessage.String())
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
	if errormessage := gjson.GetBytes(payload, `message`); errormessage.Exists() {
		panic(errormessage.String())
	}
	return nil
}
