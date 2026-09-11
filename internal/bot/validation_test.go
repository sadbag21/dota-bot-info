package bot

import (
	"errors"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestParseIDs(t *testing.T) {
	parsers := []struct {
		name     string
		commands []string
		parse    func(*tgbotapi.Message) (int64, error)
		max      string
		wantMax  int64
		overflow string
	}{
		{"Dota", []string{"player", "matches", "heroes", "stats"}, parseDotaID, "4294967295", 4294967295, "4294967296"},
		{"Match", []string{"match", "impact"}, parseMatchID, "9223372036854775807", 9223372036854775807, "9223372036854775808"},
	}
	for _, parser := range parsers {
		t.Run(parser.name, func(t *testing.T) {
			tests := []struct {
				name    string
				args    string
				want    int64
				wantErr error
			}{
				{"missing", "", 0, errMissingID},
				{"whitespace", " \t ", 0, errMissingID},
				{"minimum", "1", 1, nil},
				{"normal", "1677175114", 1677175114, nil},
				{"trimmed", " 1677175114 \t", 1677175114, nil},
				{"maximum", parser.max, parser.wantMax, nil},
				{"above maximum", parser.overflow, 0, errInvalidID},
				{"zero", "0", 0, errInvalidID},
				{"negative", "-1", 0, errInvalidID},
				{"letters", "abc", 0, errInvalidID},
				{"decimal", "1.5", 0, errInvalidID},
				{"multiple IDs", "123 456", 0, errInvalidID},
				{"trailing text", "123abc", 0, errInvalidID},
				{"integer overflow", "999999999999999999999999999", 0, errInvalidID},
			}
			if parser.name == "Dota" {
				tests = append(tests, struct {
					name, args string
					want       int64
					wantErr    error
				}{"oversized account regression", "999999999999", 0, errInvalidID})
			} else {
				tests = append(tests, struct {
					name, args string
					want       int64
					wantErr    error
				}{"match above uint32", "8983647546", 8983647546, nil})
			}
			for _, command := range parser.commands {
				for _, suffix := range []string{"", "@DotaInfoBot"} {
					t.Run(command+suffix, func(t *testing.T) {
						for _, tt := range tests {
							t.Run(tt.name, func(t *testing.T) {
								text := "/" + command + suffix
								message := &tgbotapi.Message{Text: text, Entities: []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(text)}}}
								if tt.args != "" {
									message.Text += " " + tt.args
								}
								got, err := parser.parse(message)
								if got != tt.want || !errors.Is(err, tt.wantErr) {
									t.Fatalf("parse(%q) = (%d, %v), want (%d, %v)", message.Text, got, err, tt.want, tt.wantErr)
								}
							})
						}
					})
				}
			}
		})
	}
}
