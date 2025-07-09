package irc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandInput_parseInputCharByChar(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantArgs  []string
		wantFlags map[string]string
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name:      "Normal input",
			raw:       `This is a test of some input`,
			wantArgs:  []string{`This`, `is`, `a`, `test`, `of`, `some`, `input`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Matched Double Quotes",
			raw:       `This is a "test" of some input`,
			wantArgs:  []string{`This`, `is`, `a`, `"test"`, `of`, `some`, `input`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Unmatched Double Quotes",
			raw:       `This is a test" of some input`,
			wantArgs:  []string{`This`, `is`, `a`, `test"`, `of`, `some`, `input`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Matched Single Quotes",
			raw:       `This is a 'test' of some input`,
			wantArgs:  []string{`This`, `is`, `a`, `'test'`, `of`, `some`, `input`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Unmatched Double Quotes",
			raw:       `This is a test' of some input`,
			wantArgs:  []string{`This`, `is`, `a`, `test'`, `of`, `some`, `input`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Single Word",
			raw:       `This`,
			wantArgs:  []string{`This`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Command like",
			raw:       `/me tests`,
			wantArgs:  []string{`/me`, `tests`},
			wantFlags: map[string]string{},
		},
		{
			name:      "Single flag",
			raw:       `--test=moo`,
			wantFlags: map[string]string{`test`: `moo`},
			wantArgs:  []string{},
		},
		{
			name:      "Quoted single flag, single word",
			raw:       `--test="moo"`,
			wantFlags: map[string]string{`test`: `"moo"`},
			wantArgs:  []string{},
		},
		{
			name:      "Quoted single flag, multiple word",
			raw:       `--test="moo moo"`,
			wantFlags: map[string]string{`test`: `"moo`},
			wantArgs:  []string{`moo"`},
		},
		{
			name:      "Quoted double flag and single word",
			raw:       `--test="moo" moo`,
			wantFlags: map[string]string{`test`: `"moo"`},
			wantArgs:  []string{`moo`},
		},
		{
			name:      "Quoted double flag and single word",
			raw:       `--test="moo" --text=moo`,
			wantFlags: map[string]string{`test`: `"moo"`, `text`: `moo`},
			wantArgs:  []string{},
		},
		{
			name:      "Two flags and a single word",
			raw:       `--test="moo" --text=moo moo`,
			wantFlags: map[string]string{`test`: `"moo"`, `text`: `moo`},
			wantArgs:  []string{`moo`},
		},
		{
			name:      "Two flags and a multiple words",
			raw:       `--test="moo" --text=moo moo moo`,
			wantFlags: map[string]string{`test`: `"moo"`, `text`: `moo`},
			wantArgs:  []string{`moo`, `moo`},
		},
		{
			name:      "Two flags words in middle of two flags",
			raw:       `--test="moo" moo moo --text=moo`,
			wantFlags: map[string]string{`test`: `"moo"`},
			wantArgs:  []string{`moo`, `moo`, `--text=moo`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &CommandInput{
				raw: tt.raw,
			}
			gotArgs, gotFlags, _ := ca.parseInputCharByChar(tt.raw)
			assert.Equalf(t, tt.wantArgs, gotArgs, "parseInputCharByChar() args")
			assert.Equalf(t, tt.wantFlags, gotFlags, "parseInputCharByChar() flags")
		})
	}
}
