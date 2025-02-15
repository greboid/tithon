package events

type ChannelMessageReceived struct {
	Message ChannelMessage `json:"message"`
}

type ChannelMessageSent struct {
	Message ChannelMessage `json:"message"`
}

type ChannelJoinedSelf struct {
	Channel Channel `json:"channel"`
	Time    IRCTime `json:"time"`
}

type ChannelJoinedOther struct {
	Channel Channel     `json:"channel"`
	User    ChannelUser `json:"user"`
}

type ChannelTopicChanged struct {
	Topic string      `json:"topic"`
	User  ChannelUser `json:"user"`
}

type ChannelPartedSelf struct {
	Channel Channel `json:"channel"`
}

type ChannelPartedOther struct {
	Channel Channel     `json:"channel"`
	User    ChannelUser `json:"user"`
}

type ChannelKickOther struct {
	Channel Channel     `json:"channel"`
	User    ChannelUser `json:"user"`
}

type ChannelKickSelf struct {
	Channel Channel `json:"channel"`
}

type ChannelModeChanged struct {
	Channel           Channel           `json:"channel"`
	ModeList          ModeList          `json:"modelist"`
	ModeNoParam       ModeNoParam       `json:"modenoparam"`
	ModeParamSet      ModeParamSet      `json:"modeparamset"`
	ModeParamSetUnset ModeParamSetUnset `json:"modeparamsetunset"`
}
