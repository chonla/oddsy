package oddsy

import (
	"testing"

	"github.com/nlopes/slack"
)

func TestShouldReturnDirectTypeWhenMessageIsDirectMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "D000000",
		},
	}

	result := getMessageType(ev)

	if result != DirectType {
		t.Errorf("expecting result to be %v but %v", DirectType, result)
	}
}

func TestShouldReturnPublicTypeWhenMessageIsChannelMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "C000000",
		},
	}

	result := getMessageType(ev)

	if result != PublicType {
		t.Errorf("expecting result to be %v but %v", PublicType, result)
	}
}

func TestShouldReturnUnknownTypeWhenMessageIsUnexpected(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "Q000000",
		},
	}

	result := getMessageType(ev)

	if result != UnknownType {
		t.Errorf("expecting result to be %v but %v", UnknownType, result)
	}
}

func TestShouldReturnTrueIfUIDIsFoundInMentionList(t *testing.T) {
	mentionList := []Identity{
		Identity{
			UID: "U1",
		},
		Identity{
			UID: "U2",
		},
		Identity{
			UID: "U3",
		},
	}

	result := isMentioned("U2", mentionList)

	if !result {
		t.Error("expecting result to be true but false")
	}
}

func TestShouldReturnFalseIfUIDIsNotFoundInMentionList(t *testing.T) {
	mentionList := []Identity{
		Identity{
			UID: "U1",
		},
		Identity{
			UID: "U2",
		},
		Identity{
			UID: "U3",
		},
	}

	result := isMentioned("U4", mentionList)

	if result {
		t.Error("expecting result to be false but true")
	}
}
