package init

import (
	"go.devnw.com/canary/gate"
)

const CommentPrefix = "<!--"
const CommentSuffix = "-->"

const CanaryKey = "CANARY"

const CanaryStart = "START"
const CanaryEnd = "END"

// Precomputed Start/End markers for unnamed CANARY section
var (
	// Shared markdown comment style leveraging constants
	markdownCommentStyle = gate.CommentStyle{LinePrefix: CommentPrefix, BlockEnd: CommentSuffix, Space: true}
	// Shared option chain for reuse across all updater helpers
	markdownOptions = []gate.Option{
		gate.WithStyle(markdownCommentStyle),
		gate.WithKey(CanaryKey),
		gate.WithTokens(CanaryStart, CanaryEnd),
	}
	// Public unnamed section markers (if needed externally for docs)
	StartMarker = gate.StartMarker("", markdownOptions...)
	EndMarker   = gate.EndMarker("", markdownOptions...)
)

// updateMarkdownSection updates or inserts a gated section in a markdown file
//
// The section is marked with HTML comments: <!-- CANARY:START --> ... <!-- CANARY:END -->
// If the file doesn't exist, it creates it with the content
// If the section exists, it replaces the content between the markers
// If the markers don't exist, it appends the section to the end
// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
func updateMarkdownSection(filePath, sectionContent string) error {
	return gate.UpdateSingle(filePath, sectionContent,
		append(markdownOptions, gate.WithBlankLineBefore())...,
	)
}

// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// updateMultipleMarkdownSections updates multiple gated sections in a markdown file
// Each section is identified by a unique key: <!-- CANARY:key:START --> ... <!-- CANARY:key:END -->
func updateMultipleMarkdownSections(filePath string, sections map[string]string) error {
	return gate.UpdateMultiple(filePath, sections,
		append(markdownOptions, gate.WithBlankLineBefore())...,
	)
}

// CANARY: REQ=CBIN-149; FEATURE="MarkdownSectionUpdater"; ASPECT=CLI; STATUS=IMPL; OWNER=canary; UPDATED=2025-11-01
// removeMarkdownSection removes a gated section from a markdown file
func removeMarkdownSection(filePath, sectionKey string) error {
	return gate.RemoveSection(filePath, sectionKey, markdownOptions...)
}

// buildMarkdownGatedBody returns a gated snippet (unnamed CANARY section) for direct insertion.
// Example:
// <!-- CANARY:START -->\n<content>\n<!-- CANARY:END -->\n
func buildMarkdownGatedBody(content string) string {
	return gate.BuildGatedBody(content, markdownOptions...)
}

// buildMarkdownGatedBodyKey returns a gated snippet for a keyed CANARY section.
// Example key="intro":
// <!-- CANARY:intro:START -->\n<content>\n<!-- CANARY:intro:END -->\n
func buildMarkdownGatedBodyKey(key, content string) string {
	return gate.BuildGatedBodyKey(key, content, markdownOptions...)
}
