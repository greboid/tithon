package irc

import "time"

const (
	topicTimeformat = "2006-01-02 15:04:05"
)

type Topic struct {
	topic   string
	setBy   string
	setTime time.Time
}

func NewTopic(topic, setBy string, setTime time.Time) *Topic {
	return &Topic{
		topic:   topic,
		setBy:   setBy,
		setTime: setTime,
	}
}

func (t *Topic) GetTopic() string {
	if t.topic == "" {
		return "No Topic set"
	}
	return t.topic
}

func (t *Topic) GetSetBy() string {
	return t.setBy
}

func (t *Topic) GetSetTime() time.Time {
	return t.setTime
}

func (t *Topic) GetDisplayTopic() string {
	if t.topic == "" {
		return "No Topic set"
	}

	if t.setBy != "" && !t.setTime.IsZero() {
		return t.topic + " (set by " + t.setBy + " on " + t.setTime.Format(topicTimeformat) + ")"
	}

	if t.setBy != "" {
		return t.topic + " (set by " + t.setBy + ")"
	}

	if !t.setTime.IsZero() {
		return t.topic + " (set on " + t.setTime.Format(topicTimeformat) + ")"
	}

	return t.topic
}
