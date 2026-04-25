package models

import (
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var commandTemplateVariablePattern = regexp.MustCompile(`\$\$\$([A-Za-z][A-Za-z0-9_]*)\$\$\$|\*\*\*([A-Za-z][A-Za-z0-9_]*)\*\*\*`)

type CommandTemplate struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Content   string             `bson:"content" json:"content"`
	Variables []string           `bson:"variables" json:"variables"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

func ExtractCommandTemplateVariables(content string) []string {
	matches := commandTemplateVariablePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		name := match[1]
		if name == "" {
			name = match[2]
		}
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}

	return result
}

func ReplaceCommandTemplateVariables(content string, values map[string]string) (string, []string) {
	if len(values) == 0 {
		return content, ExtractCommandTemplateVariables(content)
	}

	missing := make([]string, 0)
	seenMissing := make(map[string]struct{})
	rendered := commandTemplateVariablePattern.ReplaceAllStringFunc(content, func(token string) string {
		name := ExtractCommandTemplateVariables(token)
		if len(name) == 0 {
			return token
		}

		value, exists := values[name[0]]
		if !exists {
			if _, seen := seenMissing[name[0]]; !seen {
				seenMissing[name[0]] = struct{}{}
				missing = append(missing, name[0])
			}
			return token
		}

		return value
	})

	return rendered, missing
}
