package main

import (
	"container/list"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

var sessionDictionary = &sync.Map{}

func handleRequest(c *gin.Context) {
	bytes, _ := ioutil.ReadAll(c.Request.Body)
	if gjson.GetBytes(bytes, `object`).String() != `page` {
		c.AbortWithStatus(404)
		return
	}
	for _, entry := range gjson.GetBytes(bytes, `entry`).Array() {
		handleEvent(entry.Get(`messaging`).Array()[0])
	}
	c.Writer.WriteString(`EVENT_RECEIVED`)
	c.AbortWithStatus(200)
}
func handleEvent(messaging gjson.Result) {
	psid := messaging.Get(`sender.id`).String()
	result, found := sessionDictionary.Load(psid)
	var session *Session
	if !found {
		result, found = userList.Load(psid)
		if !found {
			initPersistentMenu(psid)
			session = &Session{
				Lock: &sync.Mutex{},
			}
			startAsking(psid, session, templates.Get(`personal`), checkInfo, func() {
				qaState := session.StateInfo.(*QAState)
				session.State = `idle`
				age, _ := strconv.Atoi(qaState.Answers[1].(string))
				userList.Store(psid, User{
					Gender:     qaState.Answers[0].(string),
					Year:       time.Now().Year() - age,
					Scam:       0,
					Unfriendly: 0,
				})
				if outro := qaState.Template.Get(`outro`); outro.Exists() {
					sendPostbackOrText(psid, outro)
				}
				session.StateInfo = nil
				changed = true
				log.Println("Just have a new user! ID: ", psid)
			}, func(oldState interface{}) {
				sendText(psid, (oldState.(*QAState)).Template.Get(`onCancel`).Value().([]interface{})...)
				sessionDictionary.Delete(psid)
			})
			return
		}
		user := result.(User)
		if checkBanned(user) {
			sendText(psid, templates.Get(`banned`).Value().([]interface{})...)
			return
		}
		session = &Session{
			State: `idle`,
			Lock:  &sync.Mutex{},
		}
		sessionDictionary.Store(psid, session)
		log.Println("New user session! ID: ", psid)
	} else {
		session = result.(*Session)
	}
	session.Lock.Lock()
	if message := messaging.Get(`message`); message.Exists() {
		if text := message.Get(`text`); text.Exists() {
			if session.State != `canceling` && strings.HasPrefix(text.String(), `#`) {
				handleCommand(psid, session, text.String())
			} else if session.State != `idle` && session.State != `finding` {
				handleText(psid, session, text.String())
			} else {
				handleCommand(psid, session, `#`+text.String())
			}
		}
		if attachments := message.Get(`attachments`); attachments.Exists() {
			if session.State == `chating` {
				handleAttachment(session, attachments)
			} else {
				sendText(psid, templates.Get(`attachmentblocking`).String())
			}
		}
	}
	if postback := messaging.Get(`postback`); postback.Exists() {
		payload := postback.Get(`payload`).String()
		switch session.State {
		case `asking`:
			if strings.HasPrefix(payload, `#`) {
				handleCommand(psid, session, payload)
				break
			}
			handleAnswer(psid, session, payload)
			break
		case `canceling`:
			handleText(psid, session, payload)
			break
		default:
			handleCommand(psid, session, payload)
			break
		}
	}
	session.Lock.Unlock()
}

func handleText(psid string, session *Session, text string) {
	switch session.State {
	case `chating`:
		sendText(session.StateInfo.(string), text)
		break
	case `asking`:
		handleAnswer(psid, session, text)
		break
	case `canceling`:
		cancelingState := session.StateInfo.(*CancelingState)
		text = strings.Trim(strings.ToLower(text), ` `)
		switch text {
		case `yes`:
			cancelingState.OnYes()
			break
		case `no`:
			cancelingState.OnNo()
			break
		default:
			sendText(psid, templates.Get(`cancel.errormessage`).String())
			break
		}
		break
	}
}
func handleAnswer(psid string, session *Session, answer string) {
	qaState := session.StateInfo.(*QAState)
	answer = strings.Trim(strings.ToLower(answer), ` `)
	if qaState.CheckFuncs[qaState.Counter](answer) {
		qaState.Answers[qaState.Counter] = answer
		qaState.Counter++
		if qaState.Counter == len(qaState.Answers) {
			qaState.OnDone()
		} else {
			sendPostbackOrText(psid, qaState.Template.Get(`questions`).Array()[qaState.Counter])
		}
	} else {
		sendText(psid, qaState.Template.Get(`questions`).Array()[qaState.Counter].Get(`errormessage`).String())
	}
}
func handleAttachment(session *Session, attachments gjson.Result) {
	for _, item := range attachments.Array() {
		sendAttachmentURL(session.StateInfo.(string), item.Get(`type`).String(), item.Get(`payload.url`).String())
	}
}
func handleCommand(psid string, session *Session, command string) {
	command = strings.ToLower(strings.Replace(command, ` `, ``, -1))
	switch command {
	case `#getstarted`:
		if session.State == `finding` || session.State == `chating` || session.State == `asking` {
			var postback Postback
			json.Unmarshal([]byte(templates.Get(`already`).Raw), &postback)
			sendPostback(psid, postback)
			return
		}
		if queue.isFull() {
			sendText(psid, templates.Get(`getstarted.onFull`).Value().([]interface{})...)
			return
		}
		startAsking(psid, session, templates.Get(`getstarted`), checkInfo, func() {
			qaState := session.StateInfo.(*QAState)
			if outro := qaState.Template.Get(`outro`); outro.Exists() {
				sendTextSync(psid, outro.Value().([]interface{})...)
			}
			session.State = `finding`
			result, _ := userList.Load(psid)
			age, _ := strconv.Atoi(qaState.Answers[1].(string))
			session.StateInfo = queue.Enqueue(&FindingRequest{
				Psid:    psid,
				Gender:  qaState.Answers[0].(string),
				Year:    time.Now().Year() - age,
				Old:     false,
				Session: session,
				User:    result.(User),
			})
			update.Signal()
			log.Println("Received a request from ID: ", psid)
		}, func(oldState interface{}) {})
		break
	case `#aboutme`:
		sendText(psid, templates.Get(`aboutme`).Value().([]interface{})...)
		break
	case `#help`:
		sendText(psid, templates.Get(`help`).Value().([]interface{})...)
		break
	case `#cancel`:
		if session.State == `idle` {
			sendText(psid, templates.Get(`nothing`).String())
			return
		}
		if session.State == `chating` {
			result, found := sessionDictionary.Load(session.StateInfo)
			if !found || (result.(*Session)).State != `chating` {
				sendText(psid, templates.Get("notsupported").String())
				return
			}
			startAsking(psid, session, templates.Get(`rating`), checkRating, func() {
				qaState := session.StateInfo.(*QAState)
				session.State = `idle`
				session.StateInfo = nil
				result, found := userList.LoadAndDelete(qaState.LastStateInfo.(string))
				if found {
					user := result.(User)
					switch qaState.Answers[0] {
					case `friendly`:
						break
					case `unfriendly`:
						user.Unfriendly++
						break
					case `scam`:
						user.Scam++
						break
					}
					userList.Store(qaState.LastStateInfo.(string), user)
					if outro := qaState.Template.Get(`outro`); outro.Exists() {
						sendText(psid, outro.Value().([]interface{})...)
					}
					changed = true

					result, found = sessionDictionary.Load(qaState.LastStateInfo)
					if found {
						othersession := result.(*Session)
						if checkBanned(user) {
							sendText(qaState.LastStateInfo.(string), templates.Get(`banned`).Value().([]interface{})...)
							othersession = nil
							sessionDictionary.Delete(qaState.LastStateInfo.(string))
						} else {
							sendTextSync(qaState.LastStateInfo.(string), templates.Get(`disconnected`).Value().([]interface{})...)
							startAsking(qaState.LastStateInfo.(string), othersession, templates.Get(`rating`), checkRating, func() {
								otherPsid := qaState.LastStateInfo.(string)
								qaState := othersession.StateInfo.(*QAState)
								othersession.State = `idle`
								othersession.StateInfo = nil

								result, found := userList.LoadAndDelete(psid)
								if found {
									user := result.(User)
									switch qaState.Answers[0] {
									case `friendly`:
										break
									case `unfriendly`:
										user.Unfriendly++
										break
									case `scam`:
										user.Scam++
										break
									}
									userList.Store(psid, user)
									if outro := qaState.Template.Get(`outro`); outro.Exists() {
										sendText(otherPsid, outro.Value().([]interface{})...)
									}
									changed = true

									if session != nil {
										if checkBanned(user) {
											sendText(psid, templates.Get(`banned`).Value().([]interface{})...)
											session = nil
											sessionDictionary.Delete(psid)
										}
									}
								}
							}, func(oldState interface{}) {
								othersession.State = `idle`
								othersession.StateInfo = nil
							})
						}
					}
				}
			}, func(oldState interface{}) {})
			return
		}
		if session.State == `asking` && (session.StateInfo.(*QAState)).Template == templates.Get(`rating`) {
			sendText(psid, templates.Get(`rating.questions`).Array()[0].Get(`errormessage`).String())
			return
		}
		cancelingState := &CancelingState{
			LastState:     session.State,
			LastStateInfo: session.StateInfo,
		}
		cancelingState.OnNo = func() {
			session.State = cancelingState.LastState
			session.StateInfo = cancelingState.LastStateInfo
		}
		cancelingState.OnYes = func() {
			switch cancelingState.LastState {
			case `asking`:
				(cancelingState.LastStateInfo.(*QAState)).OnCancel(cancelingState.LastStateInfo)
				break
			case `finding`:
				if el := cancelingState.LastStateInfo.(*list.Element); el != nil {
					go func() {
						rrLock.Lock()
						queue.Remove(el)
						sendText(psid, templates.Get(`getstarted.onCancel`).Value().([]interface{})...)
						rrLock.Unlock()
						log.Println("Canneled request of ", psid)
					}()
				}
				break
			}
			session.State = `idle`
			session.StateInfo = nil
		}
		var postback Postback
		json.Unmarshal([]byte(templates.Get(`cancel`).Raw), &postback)
		sendPostback(psid, postback)
		session.State = `canceling`
		session.StateInfo = cancelingState
		break
	default:
		sendText(psid, templates.Get(`wrongcommand.intro`).Value().([]interface{})...)
		var postback Postback
		json.Unmarshal([]byte(templates.Get(`wrongcommand.questions`).Array()[0].Raw), &postback)
		sendPostback(psid, postback)
		break
	}
}

func startAsking(psid string, session *Session, template gjson.Result, checkFuncs []func(answer string) bool,
	onDone func(), onCancel func(oldState interface{})) {
	session.State = `asking`
	qaState := &QAState{
		Template:      template,
		Answers:       make([]interface{}, len(template.Get(`questions`).Array())),
		CheckFuncs:    checkFuncs,
		Counter:       0,
		OnDone:        onDone,
		OnCancel:      onCancel,
		LastStateInfo: session.StateInfo,
	}
	session.StateInfo = qaState
	sessionDictionary.Store(psid, session)
	if intro := qaState.Template.Get(`intro`); intro.Exists() {
		sendTextSync(psid, intro.Value().([]interface{})...)
	}
	sendPostbackOrText(psid, qaState.Template.Get(`questions`).Array()[0])
}
