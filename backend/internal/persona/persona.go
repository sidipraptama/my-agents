// Package persona compiles an Agent's structured persona into a system prompt.
package persona

import (
	"fmt"
	"strings"

	"myagents/internal/models"
)

func Compile(name, role string, p models.Persona) string {
	var b strings.Builder

	fmt.Fprintf(&b, "You are %s", name)
	if role != "" {
		fmt.Fprintf(&b, ", a %s", role)
	}
	b.WriteString(".\n")

	if p.Bio != "" {
		fmt.Fprintf(&b, "\n%s\n", p.Bio)
	}
	if p.Tone != "" {
		fmt.Fprintf(&b, "\nTone and style: %s\n", p.Tone)
	}
	if len(p.Knowledge) > 0 {
		b.WriteString("\nContext you always have:\n")
		for _, k := range p.Knowledge {
			fmt.Fprintf(&b, "- %s\n", k)
		}
	}
	if len(p.Constraints) > 0 {
		b.WriteString("\nRules you must always follow:\n")
		for _, c := range p.Constraints {
			fmt.Fprintf(&b, "- %s\n", c)
		}
	}

	return b.String()
}
