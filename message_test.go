package oddsy

import (
	"testing"

	"github.com/nlopes/slack"
)

func TestGetMessageTypeOfDirectMessage(t *testing.T) {
	ev := &slack.MessageEvent{
		Msg: slack.Msg{
			Channel: "D000000",
		},
	}

	result := getMessageType(ev)

	if result != DirectType {
		t.Errorf("expecting result to be %v but %v\n", DirectType, result)
	}
}
