package tag_updater

import (
	"github.com/adamboardman/gracemount/store"
	"net/url"
	"sort"
	"strings"
)

type DisplayableTag struct {
	FirstTag string
	Tag      string
	Summary  string
}

func UpdateTags(conceptTags []store.ConceptTag, concepts []store.Concept, taggedMarkDown string, ignoreId uint) string {
	var displayedTags []DisplayableTag
	conceptMap := conceptMapFromConcepts(concepts)
	conceptTagMap := conceptTagMapFromConcepts(conceptTags)
	displayableTags := displayableTagsFromTagsMaps(conceptTags, conceptTagMap, conceptMap, ignoreId)

	var outBody strings.Builder
	displayedTags = outBodyFromMarkdownAndDisplayableTags(taggedMarkDown, displayedTags, displayableTags, &outBody)
	var outTags strings.Builder
	outTagsFromDisplayableTags(displayedTags, &outTags)

	return outTags.String() + outBody.String()
}

func outBodyFromMarkdownAndDisplayableTags(taggedMarkDown string, displayedTags []DisplayableTag, displayableTags []DisplayableTag, outBody *strings.Builder) []DisplayableTag {
	inPos := 0
	inPos = skipExistingTags(inPos, taggedMarkDown)
	for inPos < len(taggedMarkDown) {
		broken := false
		inPos, displayedTags, broken = checkAndWriteTagsToBody(displayableTags, inPos, taggedMarkDown, outBody, displayedTags, broken)
		if !broken {
			outBody.WriteString(taggedMarkDown[inPos : inPos+1])
		}
		inPos += 1
	}
	return displayedTags
}

func checkAndWriteTagsToBody(displayableTags []DisplayableTag, inPos int, taggedMarkDown string, outBody *strings.Builder, displayedTags []DisplayableTag, broken bool) (int, []DisplayableTag, bool) {
	for _, displayableTag := range displayableTags {
		tag := displayableTag.Tag
		if found, foundTag := checkForTagInMarkdown(inPos, tag, taggedMarkDown); found {
			writeTagToBody(taggedMarkDown, inPos, outBody, tag)
			displayedTags = append(displayedTags, DisplayableTag{
				FirstTag: displayableTag.FirstTag,
				Tag:      foundTag,
				Summary:  displayableTag.Summary,
			})
			inPos += len(tag) - 1
			broken = true
			break
		}
	}
	return inPos, displayedTags, broken
}

func checkForTagInMarkdown(inPos int, tag string, taggedMarkDown string) (bool, string) {
	if inPos+len(tag) < len(taggedMarkDown) {
		mdTagToCheck := taggedMarkDown[inPos : inPos+len(tag)]
		return strings.ToLower(tag) == strings.ToLower(mdTagToCheck), mdTagToCheck
	} else {
		return false, ""
	}
}

func skipExistingTags(inPos int, taggedMarkDown string) int {
	hasTags := true
	for hasTags && inPos < len(taggedMarkDown) {
		lineEndPos := strings.Index(taggedMarkDown[inPos:], "\n")
		if lineEndPos > 0 {
			line := taggedMarkDown[inPos : inPos+lineEndPos]
			endBracketPos := strings.Index(line, "]")
			colonPos := strings.Index(line, ":")
			lineEnd := line[len(line)-1:]
			if line[0:1] == "[" && endBracketPos > 2 && colonPos > endBracketPos && lineEnd != ":" {
				inPos += len(line) + 1
			} else {
				hasTags = false
			}
		} else {
			hasTags = false
		}
	}
	return inPos
}

func writeTagToBody(taggedMarkDown string, inPos int, outBody *strings.Builder, tag string) {
	needsBrackets := inPos == 0 || taggedMarkDown[inPos-1:inPos] != "["
	if needsBrackets {
		outBody.WriteString("[")
	}
	outBody.WriteString(taggedMarkDown[inPos : inPos+len(tag)])
	if needsBrackets {
		outBody.WriteString("]")
	}
}

func outTagsFromDisplayableTags(displayedTags []DisplayableTag, outTags *strings.Builder) {
	for _, displayableTag := range displayedTags {
		outTags.WriteString("[" + displayableTag.Tag + "]: /concepts/" + url.PathEscape(displayableTag.FirstTag) + " \"" + displayableTag.Summary + "\"\n")
	}
}

func displayableTagsFromTagsMaps(conceptTags []store.ConceptTag, conceptTagMap map[uint]string, conceptMap map[uint]store.Concept, ignoreId uint) []DisplayableTag {
	var displayableTags []DisplayableTag
	for _, conceptTag := range conceptTags {
		if conceptTag.ConceptId != ignoreId {
			displayableTags = append(displayableTags, DisplayableTag{
				FirstTag: conceptTagMap[conceptTag.ID],
				Tag:      conceptTag.Tag,
				Summary:  conceptMap[conceptTag.ConceptId].Summary,
			})
		}
	}
	sort.Slice(displayableTags, func(i, j int) bool {
		return len(displayableTags[i].Tag) > len(displayableTags[j].Tag)
	})
	return displayableTags
}

func conceptTagMapFromConcepts(conceptTags []store.ConceptTag) map[uint]string {
	conceptTagMap := map[uint]string{}
	for _, conceptTag := range conceptTags {
		if _, ok := conceptTagMap[conceptTag.ID]; !ok {
			conceptTagMap[conceptTag.ID] = conceptTag.Tag
		}
	}
	return conceptTagMap
}

func conceptMapFromConcepts(concepts []store.Concept) map[uint]store.Concept {
	conceptMap := map[uint]store.Concept{}
	for _, concept := range concepts {
		conceptMap[concept.ID] = concept
	}
	return conceptMap
}
