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
			want:     "dataforce",
			want1:    9,
		},
		{
			name:     "not at start, first result",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO dat",
			position: 9,
			runs:     1,
			want:     "HENLO dataforce",
			want1:    15,
		},
		{
			name:     "not at start only result, second press",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO dat",
			position: 9,
			runs:     2,
			want:     "HENLO dataforce",
			want1:    15,
		},
		{
			name:     "Two runs: First result",
			channel:  newFakeUserList("dataforce", "demented", "md87"),
			input:    "HENLO d",
			position: 7,
			runs:     2,
			want:     "HENLO demented",
			want1:    14,
		},
		{
			name:     "Three matches, three runs",
			channel:  newFakeUserList("dataforce", "demented", "dumbo"),
			input:    "HENLO d",
			position: 7,
			runs:     3,
			want:     "HENLO dumbo",
			want1:    11,
		},
		{
			name:     "No matches",
			channel:  newFakeUserList("dataforce", "demented", "dumbo"),
			input:    "HENLO z",
			position: 7,
			runs:     3,
			want:     "HENLO z",
			want1:    7,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ChannelTabCompleter{
				channel:       tt.channel,
				previousIndex: -1,
			}
			got := tt.input
			got1 := tt.position
			for _ = range tt.runs {
				got, got1 = t.Complete(got, got1)
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

func TestChannelTabCompleter_MultipleInputs(t1 *testing.T) {
	channel := newFakeUserList("dataforce", "demented", "md87")

	t := &ChannelTabCompleter{
		channel:       channel,
		previousIndex: -1,
	}

	input1, pos1 := t.Complete("d", 1)
	if input1 != "dataforce" || pos1 != 9 {
		t1.Errorf("First Complete() got `%v` at position %v, want `dataforce` at position 9", input1, pos1)
	}

	input2, pos2 := t.Complete("dataforce d", 11)
	if input2 != "dataforce dataforce" || pos2 != 19 {
		t1.Errorf("Second Complete() got `%v` at position %v, want `dataforce dataforce` at position 19", input2, pos2)
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
