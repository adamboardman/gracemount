package tag_updater

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/adamboardman/gracemount/store"
	"gorm.io/gorm"
	"os"
	"testing"
)

var s store.Store
var tags []store.ConceptTag
var concepts []store.Concept

func TestMain(m *testing.M) {
	s = store.Store{}

	s.StoreInit()

	code := m.Run()

	os.Exit(code)
}

func TestNoTagsNoChange(t *testing.T) {
	Convey("Empty String", t, func() {
		md := ""
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(md, ShouldEqual, mdOtu)
	})
	Convey("Simple text", t, func() {
		md := "Some random string with some text in it\nAnd a second line"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(md, ShouldEqual, mdOtu)
	})
	Convey("Regular web link", t, func() {
		md := "[a link](https://example.com/) and some text"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(md, ShouldEqual, mdOtu)
	})
}

func TestAddTag(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi word tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "summary",
		Full:    "",
	}}
	Convey("Add a Tag", t, func() {
		md := "Some text with a multi word tag in it"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(mdOtu, ShouldEqual, "[multi word tag]: /concepts/multi%20word%20tag \"summary\"\nSome text with a [multi word tag] in it")
	})
}

func TestUpdateTag(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi word tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "new summary",
		Full:    "",
	}}
	Convey("Update a Tag", t, func() {
		md := "[multi word tag]: /concepts/multi%20word%20tag \"summary\"\nSome text with a [multi word tag] in it"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(mdOtu, ShouldEqual, "[multi word tag]: /concepts/multi%20word%20tag \"new summary\"\nSome text with a [multi word tag] in it")
	})
}

func TestAddTagIgnoreTagCase(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi WORD tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "summary",
		Full:    "",
	}}
	Convey("Add a Tag", t, func() {
		md := "Some text with a multi word tag in it"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(mdOtu, ShouldEqual, "[multi word tag]: /concepts/multi%20WORD%20tag \"summary\"\nSome text with a [multi word tag] in it")
	})
}

func TestAddTagIgnoreMarkDownCase(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi word tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "summary",
		Full:    "",
	}}
	Convey("Add a Tag", t, func() {
		md := "Some text with a MULTI word tag in it"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(mdOtu, ShouldEqual, "[MULTI word tag]: /concepts/multi%20word%20tag \"summary\"\nSome text with a [MULTI word tag] in it")
	})
}

func TestAddTagIgnoreSelf(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi word tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "summary",
		Full:    "",
	}}
	Convey("Add a Tag", t, func() {
		md := "Some text with a MULTI word tag in it"
		mdOtu := UpdateTags(tags, concepts, md, 1)
		So(mdOtu, ShouldEqual, "Some text with a MULTI word tag in it")
	})
}

func TestUpdateTagKeepHeading(t *testing.T) {
	tags = []store.ConceptTag{store.ConceptTag{
		Model:     gorm.Model{ID: 1},
		Tag:       "multi word tag",
		ConceptId: 1,
		Order:     0,
	}}
	concepts = []store.Concept{store.Concept{
		Model:   gorm.Model{ID: 1},
		Name:    "",
		Summary: "new summary",
		Full:    "",
	}}
	Convey("Update a Tag", t, func() {
		md := "[multi word tag]: /concepts/multi%20word%20tag \"summary\"\n[TGs]:\nSome text with a [multi word tag] in it"
		mdOtu := UpdateTags(tags, concepts, md, 0)
		So(mdOtu, ShouldEqual, "[multi word tag]: /concepts/multi%20word%20tag \"new summary\"\n[TGs]:\nSome text with a [multi word tag] in it")
	})
}
