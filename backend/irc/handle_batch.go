package irc

import (
	"github.com/ergochat/irc-go/ircevent"
)

func HandleBatch() func(message *ircevent.Batch) bool {
	return func(batch *ircevent.Batch) bool {
		if len(batch.Params) > 1 && batch.Params[1] == "chathistory" {
			for i := range batch.Items {
				batch.Items[i].Message.SetTag("chathistory", "true")
			}
		}
		return false
	}
}
