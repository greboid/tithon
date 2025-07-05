package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleBatch(t *testing.T) {
	tests := []struct {
		name               string
		batch              *ircevent.Batch
		wantChathistoryTag bool
		wantReturnValue    bool
		wantItemCount      int
	}{
		{
			name: "Chathistory batch sets tags",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.Message{
							Command: "PRIVMSG",
							Params:  []string{"#channel", "Hello"},
						},
					},
					{
						Message: ircmsg.Message{
							Command: "PRIVMSG",
							Params:  []string{"#channel", "World"},
						},
					},
					{
						Message: ircmsg.Message{
							Command: "NOTICE",
							Params:  []string{"#channel", "Notice"},
						},
					},
				},
			},
			wantChathistoryTag: true,
			wantReturnValue:    false,
			wantItemCount:      3,
		},
		{
			name: "Non-chathistory batch leaves tags unchanged",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "other_type", "#channel"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.Message{
							Command: "PRIVMSG",
							Params:  []string{"#channel", "Hello"},
						},
					},
					{
						Message: ircmsg.Message{
							Command: "PRIVMSG",
							Params:  []string{"#channel", "World"},
						},
					},
				},
			},
			wantChathistoryTag: false,
			wantReturnValue:    false,
			wantItemCount:      2,
		},
		{
			name: "Empty chathistory batch",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: []*ircevent.Batch{},
			},
			wantChathistoryTag: false, // No items to tag
			wantReturnValue:    false,
			wantItemCount:      0,
		},
		{
			name: "Single item chathistory batch",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#test"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.Message{
							Command: "PRIVMSG",
							Params:  []string{"#test", "Single message"},
						},
					},
				},
			},
			wantChathistoryTag: true,
			wantReturnValue:    false,
			wantItemCount:      1,
		},
		{
			name: "Netjoin batch (non-chathistory)",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "netjoin"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.MakeMessage(nil, "", "JOIN", "#channel"),
					},
					{
						Message: ircmsg.MakeMessage(nil, "", "JOIN", "#channel"),
					},
					{
						Message: ircmsg.MakeMessage(nil, "", "JOIN", "#channel"),
					},
				},
			},
			wantChathistoryTag: false,
			wantReturnValue:    false,
			wantItemCount:      3,
		},
		{
			name: "Batch with existing tags",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: []*ircevent.Batch{
					{
						Message: func() ircmsg.Message {
							msg := ircmsg.MakeMessage(nil, "", "PRIVMSG", "#channel", "Message with existing tag")
							msg.SetTag("existing", "value")
							return msg
						}(),
					},
				},
			},
			wantChathistoryTag: true,
			wantReturnValue:    false,
			wantItemCount:      1,
		},
		{
			name: "Chathistory batch with mixed message types",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.MakeMessage(nil, "", "PRIVMSG", "#channel", "Private message"),
					},
					{
						Message: ircmsg.MakeMessage(nil, "", "NOTICE", "#channel", "Notice message"),
					},
					{
						Message: ircmsg.MakeMessage(nil, "", "JOIN", "#channel"),
					},
					{
						Message: ircmsg.MakeMessage(nil, "", "PART", "#channel", "Leaving"),
					},
				},
			},
			wantChathistoryTag: true,
			wantReturnValue:    false,
			wantItemCount:      4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := HandleBatch()

			// Execute handler
			result := handler(tt.batch)

			// Verify return value
			assert.Equal(t, tt.wantReturnValue, result, "Handler should return expected value")

			// Verify item count
			assert.Equal(t, tt.wantItemCount, len(tt.batch.Items), "Batch should have expected number of items")

			// Verify chathistory tags
			if tt.wantChathistoryTag {
				for i, item := range tt.batch.Items {
					exists, chathistoryTag := item.Message.GetTag("chathistory")
					assert.True(t, exists, "Item %d should have chathistory tag", i)
					assert.Equal(t, "true", chathistoryTag, "Item %d chathistory tag should be 'true'", i)
				}
			} else {
				for i, item := range tt.batch.Items {
					exists, _ := item.Message.GetTag("chathistory")
					assert.False(t, exists, "Item %d should not have chathistory tag", i)
				}
			}

			// Verify that other tags are preserved
			if tt.name == "Batch with existing tags" && len(tt.batch.Items) > 0 {
				exists, existingTag := tt.batch.Items[0].Message.GetTag("existing")
				assert.True(t, exists, "Existing tag should be preserved")
				assert.Equal(t, "value", existingTag, "Existing tag value should be preserved")
			}
		})
	}
}

func TestHandleBatch_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		batch           *ircevent.Batch
		wantReturnValue bool
	}{
		{
			name: "Nil batch items",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: nil,
			},
			wantReturnValue: false,
		},
		{
			name: "Batch with insufficient params",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref"}, // Missing type
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.MakeMessage(nil, "", "PRIVMSG", "#channel", "Hello"),
					},
				},
			},
			wantReturnValue: false,
		},
		{
			name: "Empty params",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.MakeMessage(nil, "", "PRIVMSG", "#channel", "Hello"),
					},
				},
			},
			wantReturnValue: false,
		},
		{
			name: "Batch with nil message",
			batch: &ircevent.Batch{
				Message: ircmsg.Message{
					Command: "BATCH",
					Params:  []string{"batch_ref", "chathistory", "#channel"},
				},
				Items: []*ircevent.Batch{
					{
						Message: ircmsg.Message{}, // Empty/default message
					},
				},
			},
			wantReturnValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandleBatch()

			// This should not panic
			result := handler(tt.batch)

			assert.Equal(t, tt.wantReturnValue, result, "Handler should return expected value")
		})
	}
}
