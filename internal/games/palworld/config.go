package palworld

import (
	"errors"
	"io"
	"os"
	"strings"
)

// readOptions reads Unreal's parenthesized OptionSettings value in the game section.
// Values remain encoded; errors must never include settings or credentials.
func readOptions(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	const maxSize = 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxSize {
		return nil, errors.New("settings file exceeds 1 MiB")
	}

	text := strings.TrimPrefix(string(data), "\ufeff")
	var options map[string]string
	var gameSection bool
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		start := offset
		offset += len(line)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			section, _, found := strings.Cut(trimmed, "]")
			if !found {
				return nil, errors.New("invalid INI section header")
			}
			gameSection = strings.EqualFold(section, "[/Script/Pal.PalGameWorldSettings")
			continue
		}
		if !gameSection {
			continue
		}
		key, _, found := strings.Cut(trimmed, "=")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "OptionSettings") {
			continue
		}
		options, err = parseOptions(text[start+strings.IndexByte(line, '=')+1:])
		if err != nil {
			return nil, err
		}
	}
	return options, nil
}

// parseOptions splits only top-level fields, respecting quoted strings and nested tuples.
func parseOptions(text string) (map[string]string, error) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "(") {
		return nil, errors.New("OptionSettings must be enclosed in parentheses")
	}
	options := make(map[string]string)
	var field strings.Builder
	depth := 1
	quoted, escaped, comment := false, false, false
	addField := func() error {
		value := strings.TrimSpace(field.String())
		field.Reset()
		if value == "" {
			return nil
		}
		key, value, found := strings.Cut(value, "=")
		if !found || strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			return errors.New("invalid field in OptionSettings")
		}
		options[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		return nil
	}
	for _, char := range text[1:] {
		if comment {
			if char == '\n' {
				comment = false
				field.WriteRune(char)
			}
			continue
		}
		if quoted {
			field.WriteRune(char)
			if escaped {
				escaped = false
			} else if char == '\\' {
				escaped = true
			} else if char == '"' {
				quoted = false
			}
			continue
		}
		switch char {
		case ';', '#':
			comment = true
			continue
		case '"':
			quoted = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return options, addField()
			}
		case ',':
			if depth == 1 {
				if err := addField(); err != nil {
					return nil, err
				}
				continue
			}
		}
		field.WriteRune(char)
	}
	return nil, errors.New("unterminated OptionSettings; check parentheses and quoted strings")
}
