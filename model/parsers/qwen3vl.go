package parsers

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"unicode"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/logutil"
)

// TODO: call the init function
const (
	CollectingThinkingContent qwenParserState = iota
	CollectingContent
	CollectingToolContent
	ThinkingDoneEatingWhitespace
	ToolCallDoneEatingWhitespace
)

const (
	thinkingCloseTag = "</think>"
	// JSON tool call patterns
	jsonToolStart = `{"name"`
	jsonToolEnd   = `}`
)

type Qwen3VLParser struct {
	state              qwenParserState
	buffer             strings.Builder
	tools              []api.Tool
	hasThinkingSupport bool
	processedToolCalls []api.ToolCall // Track tool calls found during streaming
}

func (p *Qwen3VLParser) HasToolSupport() bool {
	return true
}

func (p *Qwen3VLParser) HasThinkingSupport() bool {
	return p.hasThinkingSupport
}

func (p *Qwen3VLParser) setInitialState(lastMessage *api.Message) {
	prefill := lastMessage != nil && lastMessage.Role == "assistant"
	if !p.HasThinkingSupport() {
		p.state = CollectingContent
		return
	}

	if prefill && lastMessage.Content != "" {
		p.state = CollectingContent
		return
	}

	p.state = CollectingThinkingContent
}

func (p *Qwen3VLParser) Init(tools []api.Tool, lastMessage *api.Message) []api.Tool {
	p.tools = tools
	p.setInitialState(lastMessage)
	// Reset accumulated tool calls to prevent memory leak across requests
	p.processedToolCalls = nil
	// Reset buffer to ensure clean state
	p.buffer.Reset()
	return tools
}

type qwenEventThinkingContent struct {
	content string
}

func (qwenEventThinkingContent) isQwenEvent() {}

func (p *Qwen3VLParser) Add(s string, done bool) (content string, thinking string, calls []api.ToolCall, err error) {
	p.buffer.WriteString(s)
	
	// Debug logging removed
	
	events := p.parseEvents()

	var currentToolCalls []api.ToolCall
	var contentSb strings.Builder
	var thinkingSb strings.Builder
	for _, event := range events {
		switch event := event.(type) {
		case qwenEventRawToolCall:
			toolCall, err := parseJSONToolCall(event, p.tools)
			if err != nil {
				slog.Warn("qwen tool call parsing failed", "error", err)
				return "", "", nil, err
			}
			currentToolCalls = append(currentToolCalls, toolCall)
			// Store tool calls for final return when done=true
			p.processedToolCalls = append(p.processedToolCalls, toolCall)
		case qwenEventThinkingContent:
			thinkingSb.WriteString(event.content)
		case qwenEventContent:
			// TODO(drifkin): if the same turn contains multiple interleaved content
			// events, we naively append them together here.
			contentSb.WriteString(event.content)
		}
	}

	// When done=true, return all accumulated tool calls
	if done && len(p.processedToolCalls) > 0 {
		allToolCalls := p.processedToolCalls
		p.processedToolCalls = nil // Reset for next use
		// Debug logging removed
		return contentSb.String(), thinkingSb.String(), allToolCalls, nil
	}

	return contentSb.String(), thinkingSb.String(), currentToolCalls, nil
}

func (p *Qwen3VLParser) parseEvents() []qwenEvent {
	var all []qwenEvent

	keepLooping := true
	for keepLooping {
		var events []qwenEvent
		events, keepLooping = p.eat()
		if len(events) > 0 {
			all = append(all, events...)
		}
	}

	if len(all) > 0 {
		slog.Log(context.TODO(), logutil.LevelTrace, "qwen events parsed", "events", all, "state", p.state, "buffer", p.buffer.String())
	}

	return all
}

func splitAtTag(p *Qwen3VLParser, tag string, trimAfter bool) (string, string) {
	split := strings.SplitN(p.buffer.String(), tag, 2)
	before := split[0]
	before = strings.TrimRightFunc(before, unicode.IsSpace)
	after := split[1]
	if trimAfter {
		after = strings.TrimLeftFunc(after, unicode.IsSpace)
	}
	p.buffer.Reset()
	p.buffer.WriteString(after)
	return before, after // return events
}

func (p *Qwen3VLParser) eatLeadingWhitespaceAndTransitionTo(nextState qwenParserState) ([]qwenEvent, bool) {
	trimmed := strings.TrimLeftFunc(p.buffer.String(), unicode.IsSpace)
	p.buffer.Reset()
	if trimmed == "" {
		return nil, false
	}
	p.state = nextState
	p.buffer.WriteString(trimmed)
	return nil, true
}

func (p *Qwen3VLParser) eat() ([]qwenEvent, bool) {
	var events []qwenEvent

	switch p.state {
	case CollectingContent:
		bufferContent := p.buffer.String()
		
		// Look for XML-wrapped JSON tool calls: <tool_call>{"name": ...}</tool_call>
		if xmlStart := strings.Index(bufferContent, "<tool_call>"); xmlStart != -1 {
			before := bufferContent[:xmlStart]
			remaining := bufferContent[xmlStart:]
			
			if xmlEnd := strings.Index(remaining, "</tool_call>"); xmlEnd != -1 {
				// Complete XML-wrapped tool call found
				if len(before) > 0 {
					events = append(events, qwenEventContent{content: before})
				}
				
				// Extract JSON content between XML tags
				xmlStartLen := len("<tool_call>")
				jsonContent := remaining[xmlStartLen:xmlEnd]
				jsonContent = strings.TrimSpace(jsonContent)
				
				after := remaining[xmlEnd+len("</tool_call>"):]
				
				events = append(events, qwenEventRawToolCall{raw: jsonContent})
				p.buffer.Reset()
				p.buffer.WriteString(after)
				return events, true
			} else {
				// Incomplete XML wrapper, emit content before XML start and keep XML part
				if len(before) > 0 {
					events = append(events, qwenEventContent{content: before})
				}
				p.buffer.Reset()
				p.buffer.WriteString(remaining)
				return events, false
			}
		} else if jsonStart := strings.Index(bufferContent, jsonToolStart); jsonStart != -1 {
			// Look for direct JSON tool call pattern: {"name": "...", "arguments": {...}}
			before := bufferContent[:jsonStart]
			remaining := bufferContent[jsonStart:]
			
			// Find the end of the JSON object by counting braces
			if jsonEnd := findJSONObjectEnd(remaining); jsonEnd != -1 {
				// Complete JSON object found
				if len(before) > 0 {
					events = append(events, qwenEventContent{content: before})
				}
				
				jsonToolCall := remaining[:jsonEnd+1]
				after := remaining[jsonEnd+1:]
				
				events = append(events, qwenEventRawToolCall{raw: jsonToolCall})
				p.buffer.Reset()
				p.buffer.WriteString(after)
				return events, true
			} else {
				// Incomplete JSON object, emit content before JSON start and keep JSON part
				if len(before) > 0 {
					events = append(events, qwenEventContent{content: before})
				}
				p.buffer.Reset()
				p.buffer.WriteString(remaining)
				return events, false
			}
		} else {
			// No tool call found, emit all content except trailing whitespace
			whitespaceLen := trailingWhitespaceLen(p.buffer.String())
			ambiguousStart := len(p.buffer.String()) - whitespaceLen

			unambiguous := p.buffer.String()[:ambiguousStart]
			ambiguous := p.buffer.String()[ambiguousStart:]
			p.buffer.Reset()
			p.buffer.WriteString(ambiguous)
			if len(unambiguous) > 0 {
				events = append(events, qwenEventContent{content: unambiguous})
			}
			return events, false
		}
	case CollectingToolContent:
		// This state is not used for JSON parsing since JSON objects are parsed completely
		// Fallback to CollectingContent
		p.state = CollectingContent
		return nil, true
	case CollectingThinkingContent:
		if strings.Contains(p.buffer.String(), thinkingCloseTag) {
			thinking, remaining := splitAtTag(p, thinkingCloseTag, true)
			if len(thinking) > 0 {
				events = append(events, qwenEventThinkingContent{content: thinking})
			}
			if remaining == "" {
				p.state = ThinkingDoneEatingWhitespace
			} else {
				p.state = CollectingContent
			}
			return events, true
		} else if overlapLen := overlap(p.buffer.String(), thinkingCloseTag); overlapLen > 0 {
			beforePartialTag := p.buffer.String()[:len(p.buffer.String())-overlapLen]
			trailingWhitespaceLen := trailingWhitespaceLen(beforePartialTag)
			ambiguousStart := len(beforePartialTag) - trailingWhitespaceLen

			unambiguous := p.buffer.String()[:ambiguousStart]
			ambiguous := p.buffer.String()[ambiguousStart:]
			p.buffer.Reset()
			p.buffer.WriteString(ambiguous)
			if len(unambiguous) > 0 {
				events = append(events, qwenEventThinkingContent{content: unambiguous})
			}
			return events, false
		} else {
			whitespaceLen := trailingWhitespaceLen(p.buffer.String())
			ambiguousStart := len(p.buffer.String()) - whitespaceLen

			unambiguous := p.buffer.String()[:ambiguousStart]
			ambiguous := p.buffer.String()[ambiguousStart:]
			p.buffer.Reset()
			p.buffer.WriteString(ambiguous)
			if len(unambiguous) > 0 {
				events = append(events, qwenEventThinkingContent{content: unambiguous})
			}
			return events, false
		}
	case ThinkingDoneEatingWhitespace:
		return p.eatLeadingWhitespaceAndTransitionTo(CollectingContent)
	case ToolCallDoneEatingWhitespace:
		return p.eatLeadingWhitespaceAndTransitionTo(CollectingContent)
	default:
		panic("unreachable")
	}
}

func parseJSONToolCall(raw qwenEventRawToolCall, tools []api.Tool) (api.ToolCall, error) {
	var toolCallFunction api.ToolCallFunction
	if err := json.Unmarshal([]byte(raw.raw), &toolCallFunction); err != nil {
		return api.ToolCall{}, err
	}

	toolCall := api.ToolCall{}
	toolCall.Function = toolCallFunction

	return toolCall, nil
}

// findJSONObjectEnd finds the end of a JSON object by counting braces
// Returns the index of the closing brace, or -1 if not found
func findJSONObjectEnd(s string) int {
	braceCount := 0
	inString := false
	escaped := false
	
	for i, char := range s {
		if escaped {
			escaped = false
			continue
		}
		
		if char == '\\' {
			escaped = true
			continue
		}
		
		if char == '"' {
			inString = !inString
			continue
		}
		
		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					return i
				}
			}
		}
	}
	
	return -1 // No complete JSON object found
}
