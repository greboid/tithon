package events

import "time"

func (t *IRCTime) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(v3TimestampFormat)), nil
}

func (t *IRCTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	date, err := time.Parse(v3TimestampFormat, string(data))
	if err != nil {
		return err
	}
	*t = IRCTime{date}
	return nil
}
