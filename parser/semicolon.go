package parser

import (
	"bufio"
	"strings"
	"unicode"
)

var insertTokens map[string]struct{}

func init() {
	insertTokens = make(map[string]struct{})
	endInsertSet := "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM_1234567890)]}'\""
	for _, c := range endInsertSet {
		insertTokens[string(c)] = struct{}{}
	}
}

func insertSemis(progText string) string {
	builder := strings.Builder{}

	scanner := bufio.NewScanner(strings.NewReader(progText))
	for scanner.Scan() {
		builder.WriteString(insertLine(scanner.Text()) + "\n")
	}

	return builder.String()
}

func insertLine(line string) string {
	for i := len(line) - 1; i >= 0; i-- {
		if unicode.IsSpace(rune(line[i])) {
			continue
		}

		_, ok := insertTokens[string(line[i])]
		if ok {
			return line + ";"
		}
		break
	}

	return line
}