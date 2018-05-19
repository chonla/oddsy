package oddsy

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chonla/rnd"
	"github.com/nlopes/slack"
)

// Oddsy is slack wrapper
type Oddsy struct {
	conf   *Configuration
	api    *slack.Client
	logger *log.Logger
	rtm    *slack.RTM
	token  string
	uid    string
	Name   string
	mrFn   MessageReceivedHandlerFn
	dmrFn  DirectMessageReceivedHandlerFn
	tmrFn  map[string]FirstStringTokenReceivedHandlerFn
}

// MessageReceivedHandlerFn is message received handler function
type MessageReceivedHandlerFn func(*Oddsy, *Message)

// DirectMessageReceivedHandlerFn is direct message received handler function
type DirectMessageReceivedHandlerFn func(*Oddsy, *Message)

// FirstStringTokenReceivedHandlerFn is message received with predefined first token handler function
type FirstStringTokenReceivedHandlerFn func(*Oddsy, *Message)

// Configuration holds configuration value
type Configuration struct {
	SlackToken       string
	Debug            bool
	IgnoreBotMessage bool
}

// NewOddsy to create new oddsy
func NewOddsy(conf *Configuration) *Oddsy {
	o := &Oddsy{
		conf:   conf,
		logger: log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags),
		token:  conf.SlackToken,
		tmrFn:  map[string]FirstStringTokenReceivedHandlerFn{},
	}

	envSlackToken := os.Getenv("SLACK_TOKEN")
	if envSlackToken != "" {
		o.SetToken(envSlackToken)
	}

	return o
}

// SetToken to override token in configuration
func (o *Oddsy) SetToken(t string) {
	o.token = t
	o.logger.Println("Slack token is overwritten by environment variable.")
}

// MessageReceived hook
func (o *Oddsy) MessageReceived(h MessageReceivedHandlerFn) {
	o.mrFn = h
}

// DirectMessageReceived hook
func (o *Oddsy) DirectMessageReceived(h DirectMessageReceivedHandlerFn) {
	o.dmrFn = h
}

// FirstStringTokenReceived hook
func (o *Oddsy) FirstStringTokenReceived(t string, h FirstStringTokenReceivedHandlerFn) {
	o.tmrFn[t] = h
}

// WhoIs get user profile
func (o *Oddsy) WhoIs(id string) (u *slack.User, e error) {
	u, e = o.api.GetUserInfo(id)
	return
}

// WhatBot get bot profile
func (o *Oddsy) WhatBot(id string) (b *slack.Bot, e error) {
	b, e = o.api.GetBotInfo(id)
	return
}

// WhereIs get channel profile
func (o *Oddsy) WhereIs(id string) (c *slack.Channel, e error) {
	c, e = o.api.GetChannelInfo(id)
	return
}

// WhoAmI get bot profile
func (o *Oddsy) WhoAmI() (id string, name string) {
	id = o.uid
	name = o.Name
	return
}

// Send message
func (o *Oddsy) Send(chanID, msg string) {
	params := slack.PostMessageParameters{
		Markdown: true,
	}
	_, _, e := o.api.PostMessage(chanID, msg, params)
	if e != nil {
		o.logger.Printf("%s\n", e)
	}
}

// SendFields message
func (o *Oddsy) SendFields(chanID, msg, submsg string, values []*Field) {
	attch := slack.Attachment{
		Text:   submsg,
		Fields: []slack.AttachmentField{},
	}

	for i, n := 0, len(values); i < n; i++ {
		attch.Fields = append(attch.Fields, slack.AttachmentField{
			Title: values[i].Label,
			Value: values[i].Value,
			Short: true,
		})
	}

	params := slack.PostMessageParameters{
		Markdown:    true,
		Attachments: []slack.Attachment{attch},
	}

	_, _, e := o.api.PostMessage(chanID, msg, params)
	if e != nil {
		o.logger.Printf("%s\n", e)
	}
}

// SendSelection message
func (o *Oddsy) SendSelection(chanID, msg, submsg string, options []*SelectionOption) {
	attch := slack.Attachment{
		Text:       submsg,
		CallbackID: "tik-selections-" + rnd.Alphanum(10),
		Actions:    []slack.AttachmentAction{},
	}

	for i, n := 0, len(options); i < n; i++ {
		attch.Actions = append(attch.Actions, slack.AttachmentAction{
			Name:  options[i].Label,
			Text:  options[i].Label,
			Value: options[i].Value,
			Type:  "button",
		})
	}

	params := slack.PostMessageParameters{
		Markdown:    true,
		Attachments: []slack.Attachment{attch},
	}

	_, _, e := o.api.PostMessage(chanID, msg, params)
	if e != nil {
		o.logger.Printf("%s\n", e)
	}
}

// Start service
func (o *Oddsy) Start() {
	o.api = slack.New(o.token)
	slack.SetLogger(o.logger)
	o.api.SetDebug(o.conf.Debug)

	o.rtm = o.api.NewRTM()
	go o.rtm.ManageConnection()

	for msg := range o.rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			o.uid = ev.Info.User.ID
			o.Name = ev.Info.User.Name

		case *slack.MessageEvent:
			m := NewMessage(o, ev)
			if !m.IsBotMessage || (m.IsBotMessage && !o.conf.IgnoreBotMessage) {
				if o.mrFn != nil && m.Type == PublicType {
					o.mrFn(o, m)
				} else {
					if len(o.tmrFn) > 0 {
						ft := o.firstToken(m.Message)

						if v, ok := o.tmrFn[ft]; ok && m.Type == DirectType {
							m.Message = o.nextToken(m.Message)
							v(o, m)
						} else {
							if o.dmrFn != nil && m.Type == DirectType {
								o.dmrFn(o, m)
							}
						}
					} else {
						if o.dmrFn != nil && m.Type == DirectType {
							o.dmrFn(o, m)
						}
					}
				}
			}

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func (o *Oddsy) firstToken(t string) (r string) {
	l := strings.SplitN(t, " ", 2)
	r = l[0]
	return
}

func (o *Oddsy) nextToken(t string) (r string) {
	l := strings.SplitN(t, " ", 2)
	if len(l) > 1 {
		r = l[1]
	}
	return
}
