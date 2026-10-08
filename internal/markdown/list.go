package markdown

import (
	"regexp"
	"strings"
)

var (
	unorderedItemPattern = regexp.MustCompile(`^(\s*)[-*+]\s+(.*)$`)
	orderedItemPattern   = regexp.MustCompile(`^(\s*)\d+\.\s+(.*)$`)
	taskPattern          = regexp.MustCompile(`^\[([ xX])\]\s*(.*)$`)
)

type listLevel struct {
	tag    string
	indent int
}

func isListLine(line string) bool {
	return unorderedItemPattern.MatchString(line) || orderedItemPattern.MatchString(line)
}

func parseList(block []string) string {
	if len(block) == 0 {
		return ""
	}

	var buf strings.Builder
	stack := make([]listLevel, 0, 4)
	hasOpenLi := false

	for _, line := range block {
		var tag string
		var indent int
		var text string

		if m := unorderedItemPattern.FindStringSubmatch(line); m != nil {
			tag = "ul"
			indent = len(m[1])
			text = m[2]
		} else if m := orderedItemPattern.FindStringSubmatch(line); m != nil {
			tag = "ol"
			indent = len(m[1])
			text = m[2]
		} else {
			continue
		}

		// Check for task list syntax: [x] or [ ]
		isTask := false
		isChecked := false
		if tm := taskPattern.FindStringSubmatch(text); tm != nil {
			isTask = true
			isChecked = (tm[1] == "x" || tm[1] == "X")
			text = tm[2]
		}

		// Unwind stack for shallower indentation
		for len(stack) > 0 && indent < stack[len(stack)-1].indent {
			if hasOpenLi {
				buf.WriteString("</li>")
				hasOpenLi = false
			}
			buf.WriteString("</" + stack[len(stack)-1].tag + ">")
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				hasOpenLi = true
			}
		}

		// Change tag if indentation matches but list type changed
		if len(stack) > 0 && indent == stack[len(stack)-1].indent && tag != stack[len(stack)-1].tag {
			if hasOpenLi {
				buf.WriteString("</li>")
				hasOpenLi = false
			}
			buf.WriteString("</" + stack[len(stack)-1].tag + ">")
			stack = stack[:len(stack)-1]
		}

		// Start new nested list
		if len(stack) == 0 || indent > stack[len(stack)-1].indent {
			stack = append(stack, listLevel{tag: tag, indent: indent})
			buf.WriteString("<" + tag + ">")
			hasOpenLi = false
		} else if hasOpenLi {
			buf.WriteString("</li>")
			hasOpenLi = false
		}

		liClass := ""
		taskPrefix := ""
		if isTask {
			liClass = ` class="task-list-item"`
			if isChecked {
				taskPrefix = `<input type="checkbox" disabled checked /> `
			} else {
				taskPrefix = `<input type="checkbox" disabled /> `
			}
		}

		buf.WriteString("<li" + liClass + ">" + taskPrefix + renderInline(text))
		hasOpenLi = true
	}

	for len(stack) > 0 {
		if hasOpenLi {
			buf.WriteString("</li>")
			hasOpenLi = false
		}
		buf.WriteString("</" + stack[len(stack)-1].tag + ">")
		stack = stack[:len(stack)-1]
		if len(stack) > 0 {
			hasOpenLi = true
		}
	}

	return buf.String()
}
