package irc

type Topic struct {
	topic string
}

func NewTopic(topic string) *Topic {
	return &Topic{
		topic: topic,
	}
}

func (t *Topic) GetTopic() string {
	if t.topic == "" {
		return "No Topic set"
	}
	return t.topic
}
