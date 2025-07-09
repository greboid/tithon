package irc

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type CommandInput struct {
	raw    string
	args   []string
	flags  map[string]string
	parsed bool
}

type Argument struct {
	Name        string
	Type        ArgumentType
	Required    bool
	Default     interface{}
	Description string
	Validator   func(interface{}) error
}

type Flag struct {
	Name        string
	Short       string
	Type        ArgumentType
	Required    bool
	Default     interface{}
	Description string
	Validator   func(interface{}) error
}

type ArgumentType int

const (
	ArgTypeString ArgumentType = iota
	ArgTypeInt
	ArgTypeBool
	ArgTypeChannel
	ArgTypeNick
	ArgTypeHost
	ArgTypePort
)

type ParsedInput struct {
	args     []string
	flags    map[string]interface{}
	argNames map[string]int
}

func NewCommandInput(input string) *CommandInput {
	return &CommandInput{
		raw:    input,
		args:   make([]string, 0),
		flags:  make(map[string]string),
		parsed: false,
	}
}

func (ca *CommandInput) parseRaw() error {
	if ca.parsed {
		return nil
	}

	if strings.TrimSpace(ca.raw) == "" {
		ca.parsed = true
		return nil
	}

	args, flags, err := ca.parseInputCharByChar(ca.raw)
	if err != nil {
		return err
	}
	ca.args = args
	ca.flags = flags
	ca.parsed = true
	return nil
}

func (ca *CommandInput) parseInputCharByChar(input string) ([]string, map[string]string, error) {
	args := make([]string, 0)
	flags := make(map[string]string)
	tokens := strings.Fields(input)
	processingFlags := true
	for _, value := range tokens {
		if processingFlags && (strings.HasPrefix(value, "--") || strings.HasPrefix(value, "-") && strings.Contains(value, "=")) {
			value = strings.TrimLeft(value, "-")
			split := strings.Split(value, "=")
			flags[split[0]] = split[1]
			continue
		} else {
			processingFlags = false
		}
		args = append(args, value)
	}
	return args, flags, nil
}

func (ca *CommandInput) parseFlags(flags []Flag) (*ParsedInput, error) {
	if err := ca.parseRaw(); err != nil {
		return nil, err
	}

	result := &ParsedInput{
		args:  make([]string, 0),
		flags: make(map[string]interface{}),
	}

	result.args = ca.args

	for _, flag := range flags {
		var strValue string
		var value interface{}
		var found bool

		if flag.Type == ArgTypeBool {
			value, found = ca.flags[flag.Name]
			if !found && flag.Short != "" {
				value, found = ca.flags[flag.Short]
			}
			if !found {
				value = flag.Default
			}
			result.flags[flag.Name] = value
		} else {
			strValue, found = ca.flags[flag.Name]
			if !found && flag.Short != "" {
				strValue, found = ca.flags[flag.Short]
			}

			if found {
				var err error
				value, err = ca.parseValue(strValue, flag.Type)
				if err != nil {
					return nil, fmt.Errorf("flag %s: %w", flag.Name, err)
				}
			} else if flag.Required {
				return nil, fmt.Errorf("required flag %s is missing", flag.Name)
			} else {
				value = flag.Default
			}

			if flag.Validator != nil && value != nil {
				if err := flag.Validator(value); err != nil {
					return nil, fmt.Errorf("flag %s: %w", flag.Name, err)
				}
			}

			result.flags[flag.Name] = value
		}
	}

	return result, nil
}

func (ca *CommandInput) GetArg(index int) (string, bool) {
	if err := ca.parseRaw(); err != nil {
		return "", false
	}
	if index < 0 || index >= len(ca.args) {
		return "", false
	}
	return ca.args[index], true
}

func (ca *CommandInput) GetFlag(name string) (string, bool) {
	if err := ca.parseRaw(); err != nil {
		return "", false
	}
	value, ok := ca.flags[name]
	return value, ok
}

func (ca *CommandInput) GetBoolFlag(name string) (bool, bool) {
	if err := ca.parseRaw(); err != nil {
		return false, false
	}
	value, ok := ca.flags[name]
	if !ok {
		return false, false
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return result, true
}

func (ca *CommandInput) splitFields(input string) ([]string, error) {
	var fields []string
	var currentField strings.Builder
	var inQuotes bool
	var quoteChar rune

	for _, char := range input {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				currentField.WriteRune(char)
			}
		case ' ':
			if inQuotes {
				currentField.WriteRune(char)
			} else if currentField.Len() > 0 {
				fields = append(fields, currentField.String())
				currentField.Reset()
			}
		default:
			currentField.WriteRune(char)
		}
	}

	if inQuotes {
		return nil, errors.New("unmatched quote")
	}

	if currentField.Len() > 0 {
		fields = append(fields, currentField.String())
	}

	return fields, nil
}

func (ca *CommandInput) parseValue(value string, argType ArgumentType) (interface{}, error) {
	switch argType {
	case ArgTypeString:
		return value, nil
	case ArgTypeInt:
		return strconv.Atoi(value)
	case ArgTypeBool:
		return strconv.ParseBool(value)
	case ArgTypeChannel:
		// TODO: Should look up actual channel prefixes here
		if !strings.HasPrefix(value, "#") && !strings.HasPrefix(value, "&") {
			return nil, errors.New("channel name must start with # or &")
		}
		return value, nil
	case ArgTypeNick:
		if !isValidNick(value) {
			return nil, errors.New("invalid nickname format")
		}
		return value, nil
	case ArgTypeHost:
		if !isValidHost(value) {
			return nil, errors.New("invalid hostname format")
		}
		return value, nil
	case ArgTypePort:
		port, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("port must be a number")
		}
		if port < 1 || port > 65535 {
			return nil, errors.New("port must be between 1 and 65535")
		}
		return port, nil
	default:
		return value, nil
	}
}

func isValidNick(nick string) bool {
	if len(nick) == 0 || len(nick) > 30 {
		return false
	}
	// TODO: As with above, should actually check channel prefixes
	if strings.Contains(nick, " ") || strings.Contains(nick, "@") || strings.Contains(nick, "#") || strings.Contains(nick, "&") {
		return false
	}
	// TODO: Better nick validation ie according to RFC?
	return true
}

func isValidHost(host string) bool {
	if len(host) == 0 || len(host) > 253 {
		return false
	}

	// TODO: This appears to be a bad regex, but its the one from the validator library.  Fix and uncomment the tests
	hostRegex := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9_-]{0,62})(\.[a-zA-Z0-9_][a-zA-Z0-9_-]{0,62})*?$`)
	return hostRegex.MatchString(host)
}

func validateNonEmpty(value interface{}) error {
	if str, ok := value.(string); ok && str == "" {
		return errors.New("value cannot be empty")
	}
	return nil
}

func validateRange(min, max int) func(interface{}) error {
	return func(value interface{}) error {
		if num, ok := value.(int); ok {
			if num < min || num > max {
				return fmt.Errorf("value must be between %d and %d", min, max)
			}
		}
		return nil
	}
}

func validateOneOf(allowed []string) func(interface{}) error {
	return func(value interface{}) error {
		if str, ok := value.(string); ok {
			for _, word := range allowed {
				if str == word {
					return nil
				}
			}
			return fmt.Errorf("value must be one of: %s", strings.Join(allowed, ", "))
		}
		return nil
	}
}

func generateUsage(cmd Command) string {
	var parts []string
	parts = append(parts, cmd.GetName())

	for _, arg := range cmd.GetArgSpecs() {
		if arg.Required {
			parts = append(parts, fmt.Sprintf("<%s>", arg.Name))
		} else {
			parts = append(parts, fmt.Sprintf("[%s]", arg.Name))
		}
	}

	for _, flag := range cmd.GetFlagSpecs() {
		flagStr := "--" + flag.Name
		if flag.Short != "" {
			flagStr = fmt.Sprintf("-%s|--%s", flag.Short, flag.Name)
		}

		if flag.Type == ArgTypeBool {
			if flag.Required {
				parts = append(parts, fmt.Sprintf("<%s>", flagStr))
			} else {
				parts = append(parts, fmt.Sprintf("[%s]", flagStr))
			}
		} else {
			valueStr := fmt.Sprintf("%s=<%s>", flagStr, strings.ToLower(flag.Name))
			if flag.Required {
				parts = append(parts, fmt.Sprintf("<%s>", valueStr))
			} else {
				parts = append(parts, fmt.Sprintf("[%s]", valueStr))
			}
		}
	}

	return strings.Join(parts, " ")
}

func GenerateDetailedHelp(cmd Command) string {
	var help strings.Builder

	help.WriteString(cmd.GetHelp())
	help.WriteString("\n\nUsage: /")
	help.WriteString(generateUsage(cmd))

	// Add context information
	context := cmd.GetContext()
	if context != ContextAny {
		help.WriteString(fmt.Sprintf("\n\nContext: %s", context.String()))
	}

	argSpecs := cmd.GetArgSpecs()
	if len(argSpecs) > 0 {
		help.WriteString("\n\nArguments:")
		for _, arg := range argSpecs {
			required := ""
			if arg.Required {
				required = " (required)"
			}
			help.WriteString(fmt.Sprintf("\n  %-12s %s%s", arg.Name, arg.Description, required))
		}
	}

	flagSpecs := cmd.GetFlagSpecs()
	if len(flagSpecs) > 0 {
		help.WriteString("\n\nFlags:")
		for _, flag := range flagSpecs {
			flagName := "--" + flag.Name
			if flag.Short != "" {
				flagName = fmt.Sprintf("-%s, --%s", flag.Short, flag.Name)
			}
			required := ""
			if flag.Required {
				required = " (required)"
			}
			help.WriteString(fmt.Sprintf("\n  %-12s %s%s", flagName, flag.Description, required))
		}
	}

	return help.String()
}

func (p *ParsedInput) GetFlag(name string) (interface{}, error) {
	fmt.Println(name)
	fmt.Println(p.flags[name])
	if value, exists := p.flags[name]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("flag %s not found", name)
}

func (p *ParsedInput) GetFlagString(name string) (string, error) {
	value, err := p.GetFlag(name)
	if err != nil {
		return "", err
	}
	if str, ok := value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("unable to convert flag %s to string", name)
}

func (p *ParsedInput) GetFlagInt(name string) (int, error) {
	value, err := p.GetFlag(name)
	if err != nil {
		return -1, err
	}
	if num, ok := value.(int); ok {
		return num, nil
	}
	return -1, fmt.Errorf("unable to convert flag %s to int", name)
}

func (p *ParsedInput) GetFlagBool(name string) (bool, error) {
	if value, exists := p.flags[name]; exists {
		return value.(bool), nil
	}
	return false, fmt.Errorf("bool flag %s not found", name)
}

func (p *ParsedInput) GetArgs() ([]string, error) {
	return p.args, nil
}

func Parse(cmd Command, input string) (*ParsedInput, error) {
	ca := NewCommandInput(input)
	return ca.parseFlags(cmd.GetFlagSpecs())
}
