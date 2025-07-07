package irc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandArgs_Parse(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantArgs      []string
		wantFlags     map[string]string
		wantBoolFlags map[string]bool
		wantErr       bool
	}{
		{
			name:          "empty input",
			input:         "",
			wantArgs:      []string{},
			wantFlags:     map[string]string{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:          "simple args",
			input:         "arg1 arg2 arg3",
			wantArgs:      []string{"arg1", "arg2", "arg3"},
			wantFlags:     map[string]string{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:          "quoted args",
			input:         `"arg with spaces" 'another arg' normal`,
			wantArgs:      []string{"arg with spaces", "another arg", "normal"},
			wantFlags:     map[string]string{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:          "long flags",
			input:         "arg1 --flag1=value1 --flag2=value2",
			wantArgs:      []string{"arg1"},
			wantFlags:     map[string]string{"flag1": "value1", "flag2": "value2"},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:          "boolean flags",
			input:         "arg1 --verbose --debug",
			wantArgs:      []string{"arg1"},
			wantFlags:     map[string]string{},
			wantBoolFlags: map[string]bool{"verbose": true, "debug": true},
			wantErr:       false,
		},
		{
			name:          "mixed flags",
			input:         "arg1 --flag1=value1 --verbose -s=short arg2",
			wantArgs:      []string{"arg1", "arg2"},
			wantFlags:     map[string]string{"flag1": "value1", "s": "short"},
			wantBoolFlags: map[string]bool{"verbose": true},
			wantErr:       false,
		},
		{
			name:          "unclosed quote",
			input:         `"unclosed quote arg`,
			wantArgs:      []string{},
			wantFlags:     map[string]string{},
			wantBoolFlags: map[string]bool{},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := NewCommandInput(tt.input)
			err := ca.parseRaw()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantArgs, ca.args)
			assert.Equal(t, tt.wantFlags, ca.flags)
			assert.Equal(t, tt.wantBoolFlags, ca.boolFlags)
		})
	}
}

func TestCommandArgs_ParseWithSpecs(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		argSpecs      []Argument
		flagSpecs     []Flag
		wantArgs      []interface{}
		wantFlags     map[string]interface{}
		wantBoolFlags map[string]bool
		wantErr       bool
	}{
		{
			name:  "required string arg",
			input: "test",
			argSpecs: []Argument{
				{Name: "message", Type: ArgTypeString, Required: true},
			},
			wantArgs:      []interface{}{"test"},
			wantFlags:     map[string]interface{}{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "missing required arg",
			input: "",
			argSpecs: []Argument{
				{Name: "message", Type: ArgTypeString, Required: true},
			},
			wantErr: true,
		},
		{
			name:  "missing required arg with only spaces",
			input: "   ",
			argSpecs: []Argument{
				{Name: "message", Type: ArgTypeString, Required: true},
			},
			wantErr: true,
		},
		{
			name:  "optional arg with default",
			input: "",
			argSpecs: []Argument{
				{Name: "message", Type: ArgTypeString, Required: false, Default: "default"},
			},
			wantArgs:      []interface{}{"default"},
			wantFlags:     map[string]interface{}{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "int arg",
			input: "42",
			argSpecs: []Argument{
				{Name: "number", Type: ArgTypeInt, Required: true},
			},
			wantArgs:      []interface{}{42},
			wantFlags:     map[string]interface{}{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "invalid int arg",
			input: "notanumber",
			argSpecs: []Argument{
				{Name: "number", Type: ArgTypeInt, Required: true},
			},
			wantErr: true,
		},
		{
			name:  "channel arg",
			input: "#test",
			argSpecs: []Argument{
				{Name: "channel", Type: ArgTypeChannel, Required: true},
			},
			wantArgs:      []interface{}{"#test"},
			wantFlags:     map[string]interface{}{},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "invalid channel arg",
			input: "test",
			argSpecs: []Argument{
				{Name: "channel", Type: ArgTypeChannel, Required: true},
			},
			wantErr: true,
		},
		{
			name:  "boolean flag",
			input: "--verbose",
			flagSpecs: []Flag{
				{Name: "verbose", Type: ArgTypeBool, Default: false},
			},
			wantArgs:      []interface{}{},
			wantFlags:     map[string]interface{}{},
			wantBoolFlags: map[string]bool{"verbose": true},
			wantErr:       false,
		},
		{
			name:  "string flag",
			input: "--password=secret",
			flagSpecs: []Flag{
				{Name: "password", Type: ArgTypeString, Required: true},
			},
			wantArgs:      []interface{}{},
			wantFlags:     map[string]interface{}{"password": "secret"},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "port flag",
			input: "--port=6667",
			flagSpecs: []Flag{
				{Name: "port", Type: ArgTypePort, Default: 6667},
			},
			wantArgs:      []interface{}{},
			wantFlags:     map[string]interface{}{"port": 6667},
			wantBoolFlags: map[string]bool{},
			wantErr:       false,
		},
		{
			name:  "invalid port flag",
			input: "--port=70000",
			flagSpecs: []Flag{
				{Name: "port", Type: ArgTypePort, Default: 6667},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := NewCommandInput(tt.input)
			result, err := ca.parseFlags(tt.argSpecs, tt.flagSpecs)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantArgs, result.args)
			assert.Equal(t, tt.wantFlags, result.flags)
			assert.Equal(t, tt.wantBoolFlags, result.boolFlags)
		})
	}
}

func TestCommandArgs_GetArg(t *testing.T) {
	ca := NewCommandInput("arg1 arg2 arg3")

	arg, ok := ca.GetArg(0)
	assert.True(t, ok)
	assert.Equal(t, "arg1", arg)

	arg, ok = ca.GetArg(1)
	assert.True(t, ok)
	assert.Equal(t, "arg2", arg)

	arg, ok = ca.GetArg(2)
	assert.True(t, ok)
	assert.Equal(t, "arg3", arg)

	arg, ok = ca.GetArg(3)
	assert.False(t, ok)
	assert.Equal(t, "", arg)

	arg, ok = ca.GetArg(-1)
	assert.False(t, ok)
	assert.Equal(t, "", arg)
}

func TestCommandArgs_GetFlag(t *testing.T) {
	ca := NewCommandInput("--flag1=value1 --flag2=value2")

	value, ok := ca.GetFlag("flag1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	value, ok = ca.GetFlag("flag2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)

	value, ok = ca.GetFlag("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestCommandArgs_GetBoolFlag(t *testing.T) {
	ca := NewCommandInput("--verbose --debug")

	assert.True(t, ca.GetBoolFlag("verbose"))
	assert.True(t, ca.GetBoolFlag("debug"))
	assert.False(t, ca.GetBoolFlag("nonexistent"))
}

func TestValidateNonEmpty(t *testing.T) {
	assert.NoError(t, validateNonEmpty("test"))
	assert.Error(t, validateNonEmpty(""))
	assert.NoError(t, validateNonEmpty(123))
}

func TestValidateRange(t *testing.T) {
	validator := validateRange(1, 10)

	assert.NoError(t, validator(5))
	assert.NoError(t, validator(1))
	assert.NoError(t, validator(10))
	assert.Error(t, validator(0))
	assert.Error(t, validator(11))
	assert.NoError(t, validator("string"))
}

func TestValidateOneOf(t *testing.T) {
	validator := validateOneOf([]string{"red", "green", "blue"})

	assert.NoError(t, validator("red"))
	assert.NoError(t, validator("green"))
	assert.NoError(t, validator("blue"))
	assert.Error(t, validator("yellow"))
	assert.NoError(t, validator(123))
}

func TestIsValidNick(t *testing.T) {
	tests := []struct {
		nick  string
		valid bool
	}{
		{"test", true},
		{"Test123", true},
		{"test_user", true},
		{"[bot]", true},
		{"{user}", true},
		{"user|away", true},
		{"user^", true},
		{"user`", true},
		{"user\\", true},
		{"", false},
		{"user with spaces", false},
		{"user@host", false},
		{"user#tag", false},
	}

	for _, tt := range tests {
		t.Run(tt.nick, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidNick(tt.nick))
		})
	}
}

func TestIsValidHost(t *testing.T) {
	tests := []struct {
		host  string
		valid bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"localhost", true},
		{"192.168.1.1", true},
		{"", false},
		// {"example-.com", false},
		// {"-example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidHost(tt.host))
		})
	}
}

func TestGenerateUsage(t *testing.T) {
	cmd := &AddServer{}
	usage := generateUsage(cmd)

	expected := "addserver <hostname> [nickname] [--notls] [-p|--password=<password>] [-s|--sasl=<sasl>]"
	assert.Equal(t, expected, usage)
}

func TestGenerateDetailedHelp(t *testing.T) {
	cmd := &AddServer{}
	help := GenerateDetailedHelp(cmd)

	assert.Contains(t, help, "Adds a new server and connects to it")

	assert.Contains(t, help, "Usage: /addserver")

	assert.Contains(t, help, "Arguments:")
	assert.Contains(t, help, "hostname")
	assert.Contains(t, help, "(required)")

	assert.Contains(t, help, "Flags:")
	assert.Contains(t, help, "-p, --password")
	assert.Contains(t, help, "Disable TLS encryption")
}

func TestParseCommandWithSpecs(t *testing.T) {
	cmd := &AddServer{}

	parsed, err := Parse(cmd, "irc.example.com --notls -p=secret")
	assert.NoError(t, err)
	assert.Equal(t, "irc.example.com", parsed.args[0])
	assert.True(t, parsed.boolFlags["notls"])
	assert.Equal(t, "secret", parsed.flags["password"])

	_, err = Parse(cmd, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required argument hostname is missing")
}

func TestParsedInput_GetArgByName(t *testing.T) {
	argSpecs := []Argument{
		{Name: "hostname", Type: ArgTypeString, Required: true},
		{Name: "port", Type: ArgTypeInt, Required: false, Default: 6667},
		{Name: "ssl", Type: ArgTypeBool, Required: false, Default: false},
	}

	ca := NewCommandInput("irc.example.com 6697 true")
	parsed, err := ca.parseFlags(argSpecs, []Flag{})
	assert.NoError(t, err)

	// Test GetArg
	hostname, err := parsed.GetArg("hostname")
	assert.Nil(t, err)
	assert.Equal(t, "irc.example.com", hostname)

	port, err := parsed.GetArg("port")
	assert.Nil(t, err)
	assert.Equal(t, 6697, port)

	ssl, err := parsed.GetArg("ssl")
	assert.Nil(t, err)
	assert.Equal(t, true, ssl)

	// Test non-existent arg
	_, err = parsed.GetArg("nonexistent")
	assert.Error(t, err)

	// Test GetArgString
	hostnameStr, err := parsed.GetArgString("hostname")
	assert.Nil(t, err)
	assert.Equal(t, "irc.example.com", hostnameStr)

	// Test GetArgString with wrong type
	_, err = parsed.GetArgString("port")
	assert.Error(t, err)

	// Test GetArgInt
	portInt, err := parsed.GetArgInt("port")
	assert.Nil(t, err)
	assert.Equal(t, 6697, portInt)

	// Test GetArgInt with wrong type
	_, err = parsed.GetArgInt("hostname")
	assert.Error(t, err)

	// Test GetArgBool
	sslBool, err := parsed.GetArgBool("ssl")
	assert.Nil(t, err)
	assert.Equal(t, true, sslBool)

	// Test GetArgBool with wrong type
	_, err = parsed.GetArgBool("hostname")
	assert.Error(t, err)
}

func TestParsedInput_GetArgByName_WithDefaults(t *testing.T) {
	argSpecs := []Argument{
		{Name: "hostname", Type: ArgTypeString, Required: true},
		{Name: "port", Type: ArgTypeInt, Required: false, Default: 6667},
		{Name: "nickname", Type: ArgTypeString, Required: false, Default: "defaultnick"},
	}

	// Test with only required argument provided
	ca := NewCommandInput("irc.example.com")
	parsed, err := ca.parseFlags(argSpecs, []Flag{})
	assert.NoError(t, err)

	// Test that all arguments are accessible by name, including defaults
	hostname, err := parsed.GetArgString("hostname")
	assert.NoError(t, err)
	assert.Equal(t, "irc.example.com", hostname)

	port, err := parsed.GetArgInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 6667, port)

	nickname, err := parsed.GetArgString("nickname")
	assert.NoError(t, err)
	assert.Equal(t, "defaultnick", nickname)

	// Test that args slice also contains the defaults
	assert.Equal(t, "irc.example.com", parsed.args[0])
	assert.Equal(t, 6667, parsed.args[1])
	assert.Equal(t, "defaultnick", parsed.args[2])

	// Test with partial arguments provided
	ca2 := NewCommandInput("irc.example.com 8080")
	parsed2, err := ca2.parseFlags(argSpecs, []Flag{})
	assert.NoError(t, err)

	hostname2, _ := parsed2.GetArgString("hostname")
	assert.Equal(t, "irc.example.com", hostname2)

	port2, _ := parsed2.GetArgInt("port")
	assert.Equal(t, 8080, port2)

	nickname2, _ := parsed2.GetArgString("nickname")
	assert.Equal(t, "defaultnick", nickname2)
}

func TestParsedInput_GetFlagByName(t *testing.T) {
	flagSpecs := []Flag{
		{Name: "password", Type: ArgTypeString, Required: false, Default: ""},
		{Name: "port", Type: ArgTypePort, Required: false, Default: 6667},
		{Name: "verbose", Type: ArgTypeBool, Required: false, Default: false},
	}

	ca := NewCommandInput("--password=secret --port=8080 --verbose")
	parsed, err := ca.parseFlags([]Argument{}, flagSpecs)
	assert.NoError(t, err)

	// Test GetFlag
	password, err := parsed.GetFlag("password")
	assert.NoError(t, err)
	assert.Equal(t, "secret", password)

	port, err := parsed.GetFlag("port")
	assert.NoError(t, err)
	assert.Equal(t, 8080, port)

	// Test non-existent flag
	_, err = parsed.GetFlag("nonexistent")
	assert.Error(t, err)

	// Test GetFlagString
	passwordStr, err := parsed.GetFlagString("password")
	assert.NoError(t, err)
	assert.Equal(t, "secret", passwordStr)

	// Test GetFlagString with wrong type
	_, err = parsed.GetFlagString("port")
	assert.Error(t, err)

	// Test GetFlagInt
	portInt, err := parsed.GetFlagInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 8080, portInt)

	// Test GetFlagInt with wrong type
	_, err = parsed.GetFlagInt("password")
	assert.Error(t, err)

	// Test GetFlagBool
	verboseBool, err := parsed.GetFlagBool("verbose")
	assert.NoError(t, err)
	assert.Equal(t, true, verboseBool)

	// Test GetFlagBool with non-existent flag
	_, err = parsed.GetFlagBool("nonexistent")
	assert.Error(t, err)
}

func TestParsedInput_GetFlagByName_WithDefaults(t *testing.T) {
	flagSpecs := []Flag{
		{Name: "password", Type: ArgTypeString, Required: false, Default: "defaultpass"},
		{Name: "port", Type: ArgTypePort, Required: false, Default: 6667},
		{Name: "verbose", Type: ArgTypeBool, Required: false, Default: false},
	}

	// Test with no flags provided - should use defaults
	ca := NewCommandInput("")
	parsed, err := ca.parseFlags([]Argument{}, flagSpecs)
	assert.NoError(t, err)

	// Test that all flags are accessible by name, including defaults
	password, err := parsed.GetFlagString("password")
	assert.NoError(t, err)
	assert.Equal(t, "defaultpass", password)

	port, err := parsed.GetFlagInt("port")
	assert.NoError(t, err)
	assert.Equal(t, 6667, port)

	verbose, err := parsed.GetFlagBool("verbose")
	assert.NoError(t, err)
	assert.Equal(t, false, verbose)

	// Test that flags and boolFlags maps also contain the defaults
	assert.Equal(t, "defaultpass", parsed.flags["password"])
	assert.Equal(t, 6667, parsed.flags["port"])
	assert.Equal(t, false, parsed.boolFlags["verbose"])
}

func TestArgTypeRestOfInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single word",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "Multiple words",
			input:    "Hello world this is a test",
			expected: "Hello world this is a test",
		},
		{
			name:     "Input with extra spaces",
			input:    "Hello   world   test",
			expected: "Hello world test",
		},
		{
			name:     "Quoted strings within input",
			input:    "Hello \"quoted text\" and more",
			expected: "Hello quoted text and more",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create command args with RestOfInput type
			args := []Argument{
				{
					Name:        "message",
					Type:        ArgTypeRestOfInput,
					Required:    true,
					Description: "Test message",
				},
			}

			ca := NewCommandInput(tt.input)
			result, err := ca.parseFlags(args, []Flag{})

			assert.NoError(t, err)
			message, err := result.GetArgString("message")
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, message)
		})
	}
}

func TestArgTypeRestOfInputWithMultipleArgs(t *testing.T) {
	// Test case where RestOfInput is the second argument
	args := []Argument{
		{
			Name:        "target",
			Type:        ArgTypeNick,
			Required:    true,
			Description: "Target nickname",
		},
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
			Required:    true,
			Description: "Message text",
		},
	}

	ca := NewCommandInput("alice Hello world this is a message")
	result, err := ca.parseFlags(args, []Flag{})

	assert.NoError(t, err)
	
	target, err := result.GetArgString("target")
	assert.NoError(t, err)
	assert.Equal(t, "alice", target)
	
	message, err := result.GetArgString("message")
	assert.NoError(t, err)
	assert.Equal(t, "Hello world this is a message", message)
}

func TestArgTypeRestOfInputOptional(t *testing.T) {
	// Test optional RestOfInput argument
	args := []Argument{
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
			Required:    false,
			Default:     "default message",
			Description: "Optional message",
		},
	}

	// Test with empty input
	ca := NewCommandInput("")
	result, err := ca.parseFlags(args, []Flag{})

	assert.NoError(t, err)
	message, err := result.GetArgString("message")
	assert.NoError(t, err)
	assert.Equal(t, "default message", message)

	// Test with actual input
	ca2 := NewCommandInput("actual message text")
	result2, err := ca2.parseFlags(args, []Flag{})

	assert.NoError(t, err)
	message2, err := result2.GetArgString("message")
	assert.NoError(t, err)
	assert.Equal(t, "actual message text", message2)
}

func TestArgTypeRestOfInputEmptyRequired(t *testing.T) {
	// Test required RestOfInput argument with empty input
	args := []Argument{
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
			Required:    true,
			Description: "Required message",
		},
	}

	ca := NewCommandInput("")
	_, err := ca.parseFlags(args, []Flag{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required argument message is missing")
}
