package oddsy

import (
	"regexp"

	"github.com/nlopes/slack"
)

// Message is wrapped message
type Message struct {
	From         Identity
	IsBotMessage bool
	Message      string
	Channel      Identity
	Type         MessageType
	Mentioned    bool
	MentionList  []Identity
}

// MessageType is type of slack message
type MessageType int

const (
	// PublicType is public message
	PublicType MessageType = iota
	// DirectType is public message
	DirectType
	// BotType is message from bot
	BotType
	// UnknownType is unexpected message type
	UnknownType
)

var userReg = regexp.MustCompile("<@(U[^>]+)>")

// NewMessage parses slack message and wrap it
func NewMessage(o I, ev *slack.MessageEvent) *Message {
	var uName string
	var uID string
	isBot := (ev.SubType == "bot_message")
	if isBot {
		uID = ev.BotID
		b, e := o.WhatBot(ev.BotID)
		if e == nil {
			uName = b.Name
		}
	} else {
		uID = ev.User
		u, e := o.WhoIs(ev.User)
		if e == nil {
			uName = u.Name
		} else {
			uName = "?"
		}
	}
	mentions := getMentionList(o, ev.Text)
	mentioned := isMentioned(o.UID(), mentions)

	m := &Message{
		From: Identity{
			Name: uName,
			UID:  uID,
		},
		Message: ev.Text,
		Channel: Identity{
			UID: ev.Channel,
		},
		MentionList:  mentions,
		Mentioned:    mentioned,
		IsBotMessage: isBot,
	}

	switch getType(ev) {
	case DirectType:
		m.Channel.Name = "Direct Message"
		m.Type = DirectType
	case PublicType:
		m.Type = PublicType
		c, e := o.WhereIs(ev.Channel)
		if e == nil {
			m.Channel.Name = c.Name
		}
	case UnknownType:
		m.Type = UnknownType
		c, e := o.WhereIs(ev.Channel)
		if e == nil {
			m.Channel.Name = c.Name
		}
	}
	return m
}

func getMentionList(o I, m string) []Identity {
	list := userReg.FindAllStringSubmatch(m, -1)
	ids := []Identity{}
	for i, n := 0, len(list); i < n; i++ {
		u, _ := o.WhoIs(list[i][1])
		ids = append(ids, Identity{
			Name: u.Name,
			UID:  u.ID,
		})
	}
	return ids
}

func isMentioned(id string, l []Identity) bool {
	for i, n := 0, len(l); i < n; i++ {
		if id == l[i].UID {
			return true
		}
	}
	return false
}

func getType(ev *slack.MessageEvent) MessageType {
	switch ev.Channel[0:1] {
	case "D":
		return DirectType
	case "C":
		return PublicType
	}
	return UnknownType
}
