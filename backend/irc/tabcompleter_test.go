package irc

import (
	"slices"
	"strings"
	"testing"
)

func TestChannelTabCompleter_Complete(t1 *testing.T) {
	tests := []struct {
		name     string
		channel  userList
		input    string
		position int
		runs     int
		want     string
		want1    int
	}{
		{
			name:     "Start of text, first result",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "dat",
			position: 3,
			runs:     1,
			want:     "dataforce ",
			want1:    9,
		},
		{
			name:     "not at start, first result",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO dat",
			position: 9,
			runs:     1,
			want:     "HENLO dataforce ",
			want1:    15,
		},
		{
			name:     "not at start only result, second press",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO dat",
			position: 9,
			runs:     2,
			want:     "HENLO dataforce ",
			want1:    15,
		},
		{
			name:     "Two runs: First result",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO d",
			position: 7,
			runs:     2,
			want:     "HENLO demented ",
			want1:    15,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ChannelTabCompleter{
				channel:       tt.channel,
				previousIndex: -1,
			}
			var got string
			var got1 int
			for _ = range tt.runs {
				got, got1 = t.Complete(tt.input, tt.position)
			}
			if got != tt.want {
				t1.Errorf("Complete() got = `%v`, want `%v`", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("Complete() got1 = `%v`, want `%v`", got1, tt.want1)
			}
		})
	}
}

type fakeUserListGetter struct {
	users []*User
}

func newFakeUserList(users ...string) *fakeUserListGetter {
	ful := &fakeUserListGetter{}
	for i := range users {
		ful.users = append(ful.users, NewUser(users[i], ""))
	}
	slices.SortFunc(ful.users, func(a, b *User) int {
		modeCmp := strings.Compare(b.modes, a.modes)
		if modeCmp != 0 {
			return modeCmp
		}
		return strings.Compare(a.nickname, b.nickname)
	})
	return ful
}

func (f fakeUserListGetter) GetUsers() []*User {
	return f.users
}
