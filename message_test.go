package oddsy_test

import (
	"errors"
	"testing"

	"github.com/chonla/oddsy"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedOddsy struct {
	mock.Mock
}

func (o *MockedOddsy) WhoIs(id string) (*slack.User, error) {
	args := o.Called(id)
	return args.Get(0).(*slack.User), args.Error(1)
}

func (o *MockedOddsy) UID() string {
	args := o.Called()
	return args.String(0)
}

func (o *MockedOddsy) WhatBot(id string) (*slack.Bot, error) {
	args := o.Called(id)
	return args.Get(0).(*slack.Bot), args.Error(1)
}

func (o *MockedOddsy) WhereIs(id string) (*slack.Channel, error) {
	args := o.Called(id)
	return args.Get(0).(*slack.Channel), args.Error(1)
}

func (o *MockedOddsy) WhoAmI() (string, string) {
	args := o.Called()
	return args.String(0), args.String(1)
}

func TestShouldCreateDirectTypeMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "D000000",
			User:    "U1111",
			Text:    "Some Message",
		},
	}
	expected := &oddsy.Message{
		From: oddsy.Identity{
			UID:  ev.Msg.User,
			Name: "Some Name",
		},
		IsBotMessage: false,
		Message:      ev.Msg.Text,
		Channel: oddsy.Identity{
			UID:  ev.Msg.Channel,
			Name: "Direct Message",
		},
		Type:        oddsy.DirectType,
		Mentioned:   false,
		MentionList: []oddsy.Identity{},
	}

	oddsyMock := new(MockedOddsy)
	oddsyMock.On("UID").Return("U0000")
	oddsyMock.On("WhoIs", ev.Msg.User).Return(&slack.User{
		Name: "Some Name",
		ID:   ev.Msg.User,
	}, nil)
	oddsyMock.On("WhereIs", ev.Msg.Channel).Return(&slack.Channel{}, errors.New("Don't worry. I just don't know how to mock private things."))

	m := oddsy.NewMessage(oddsyMock, ev)

	assert.Equal(t, expected, m)
}

func TestShouldCreatePublicTypeMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "C000000",
			User:    "U1111",
			Text:    "Some Message",
		},
	}
	expected := &oddsy.Message{
		From: oddsy.Identity{
			UID:  ev.Msg.User,
			Name: "Some Name",
		},
		IsBotMessage: false,
		Message:      ev.Msg.Text,
		Channel: oddsy.Identity{
			UID:  ev.Msg.Channel,
			Name: "",
		},
		Type:        oddsy.PublicType,
		Mentioned:   false,
		MentionList: []oddsy.Identity{},
	}

	oddsyMock := new(MockedOddsy)
	oddsyMock.On("UID").Return("U0000")
	oddsyMock.On("WhoIs", ev.Msg.User).Return(&slack.User{
		Name: "Some Name",
		ID:   ev.Msg.User,
	}, nil)
	oddsyMock.On("WhereIs", ev.Msg.Channel).Return(&slack.Channel{}, errors.New("Don't worry. I just don't know how to mock private things."))

	m := oddsy.NewMessage(oddsyMock, ev)

	assert.Equal(t, expected, m)
}

func TestShouldCreateUnknownTypeMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "Z000000",
			User:    "U1111",
			Text:    "Some Message",
		},
	}
	expected := &oddsy.Message{
		From: oddsy.Identity{
			UID:  ev.Msg.User,
			Name: "Some Name",
		},
		IsBotMessage: false,
		Message:      ev.Msg.Text,
		Channel: oddsy.Identity{
			UID:  ev.Msg.Channel,
			Name: "",
		},
		Type:        oddsy.UnknownType,
		Mentioned:   false,
		MentionList: []oddsy.Identity{},
	}

	oddsyMock := new(MockedOddsy)
	oddsyMock.On("UID").Return("U0000")
	oddsyMock.On("WhoIs", ev.Msg.User).Return(&slack.User{
		Name: "Some Name",
		ID:   ev.Msg.User,
	}, nil)
	oddsyMock.On("WhereIs", ev.Msg.Channel).Return(&slack.Channel{}, errors.New("Don't worry. I just don't know how to mock private things."))

	m := oddsy.NewMessage(oddsyMock, ev)

	assert.Equal(t, expected, m)
}

func TestShouldCreateMentionListCorrectly(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "C000000",
			User:    "U1111",
			Text:    "Some Message <@U0000> <@U1112> <@U1113>",
		},
	}
	expected := &oddsy.Message{
		From: oddsy.Identity{
			UID:  ev.Msg.User,
			Name: "Some Name",
		},
		IsBotMessage: false,
		Message:      ev.Msg.Text,
		Channel: oddsy.Identity{
			UID:  ev.Msg.Channel,
			Name: "",
		},
		Type:      oddsy.PublicType,
		Mentioned: true,
		MentionList: []oddsy.Identity{
			oddsy.Identity{
				UID:  "U0000",
				Name: "Bot Name",
			},
			oddsy.Identity{
				UID:  "U1112",
				Name: "Some Name 1",
			},
			oddsy.Identity{
				UID:  "U1113",
				Name: "Some Name 2",
			},
		},
	}

	oddsyMock := new(MockedOddsy)
	oddsyMock.On("UID").Return("U0000")
	oddsyMock.On("WhoIs", ev.Msg.User).Return(&slack.User{
		Name: "Some Name",
		ID:   ev.Msg.User,
	}, nil)
	oddsyMock.On("WhoIs", "U0000").Return(&slack.User{
		Name: "Bot Name",
		ID:   "U0000",
	}, nil)
	oddsyMock.On("WhoIs", "U1112").Return(&slack.User{
		Name: "Some Name 1",
		ID:   "U1112",
	}, nil)
	oddsyMock.On("WhoIs", "U1113").Return(&slack.User{
		Name: "Some Name 2",
		ID:   "U1113",
	}, nil)
	oddsyMock.On("WhereIs", ev.Msg.Channel).Return(&slack.Channel{}, errors.New("Don't worry. I just don't know how to mock private things."))

	m := oddsy.NewMessage(oddsyMock, ev)

	assert.Equal(t, expected, m)
}

func TestShouldNotBeMentionedWhenNoBotIDInMentionList(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "C000000",
			User:    "U1111",
			Text:    "Some Message <@U1112> <@U1113>",
		},
	}
	expected := &oddsy.Message{
		From: oddsy.Identity{
			UID:  ev.Msg.User,
			Name: "Some Name",
		},
		IsBotMessage: false,
		Message:      ev.Msg.Text,
		Channel: oddsy.Identity{
			UID:  ev.Msg.Channel,
			Name: "",
		},
		Type:      oddsy.PublicType,
		Mentioned: false,
		MentionList: []oddsy.Identity{
			oddsy.Identity{
				UID:  "U1112",
				Name: "Some Name 1",
			},
			oddsy.Identity{
				UID:  "U1113",
				Name: "Some Name 2",
			},
		},
	}

	oddsyMock := new(MockedOddsy)
	oddsyMock.On("UID").Return("U0000")
	oddsyMock.On("WhoIs", ev.Msg.User).Return(&slack.User{
		Name: "Some Name",
		ID:   ev.Msg.User,
	}, nil)
	oddsyMock.On("WhoIs", "U1112").Return(&slack.User{
		Name: "Some Name 1",
		ID:   "U1112",
	}, nil)
	oddsyMock.On("WhoIs", "U1113").Return(&slack.User{
		Name: "Some Name 2",
		ID:   "U1113",
	}, nil)
	oddsyMock.On("WhereIs", ev.Msg.Channel).Return(&slack.Channel{}, errors.New("Don't worry. I just don't know how to mock private things."))

	m := oddsy.NewMessage(oddsyMock, ev)

	assert.Equal(t, expected, m)
}
