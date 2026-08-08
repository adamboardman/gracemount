package store

import (
	"log"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

var s Store

func TestMain(m *testing.M) {
	s = Store{}
	s.StoreInit()

	code := m.Run()

	os.Exit(code)
}

func ensureTestUserExists(emailAddress string, permissions UserPermissions) *User {
	user, err := s.FindUser(emailAddress)
	if err != nil {
		encrypted, err := bcrypt.GenerateFromPassword([]byte("1234"), 13)
		So(err, ShouldBeNil)
		user = &User{
			Password: string(encrypted),
			PrivilegedUser: PrivilegedUser{
				PublicUser: PublicUser{
					Email: emailAddress,
				},
				Permissions: permissions,
				Confirmed:   true,
			},
		}
		_, _ = s.InsertUser(user)
	} else if user.Permissions != permissions {
		user.Permissions = permissions
		s.UpdateUser(user)
	}
	return user
}

func TestStore_DoubleInsertUser(t *testing.T) {
	const emailAddress = "joe@example.com"
	Convey("Insert a user to the store", t, func() {
		s.PurgeUser(emailAddress)
		user := User{}
		user.Email = emailAddress
		user.FirstName = "Joe"
		user.LastName = "Blogs"
		userId, _ := s.InsertUser(&user)

		Convey("User should be given an ID", func() {
			So(userId, ShouldBeGreaterThan, 0)
		})

		Convey("Insert the same email again", func() {
			user2 := User{}
			user2.Email = emailAddress
			user2.FirstName = "John"
			user2.LastName = "Smith"
			user2Id, err := s.InsertUser(&user2)

			Convey("Expect Error and No UserId", func() {
				So(err, ShouldNotBeNil)
				So(user2Id, ShouldEqual, 0)
			})
		})

	})
}

func TestStore_InsertConcept(t *testing.T) {
	const name = "test"
	Convey("Insert a concept to the store", t, func() {
		s.PurgeConcept(name)
		concept := Concept{Name: name}
		concept.Summary = "a short version of the test concept"
		conceptId, _ := s.InsertConcept(&concept)

		Convey("Concept should be given an ID", func() {
			So(conceptId, ShouldBeGreaterThan, 0)
		})

		Convey("Concepts list should contain concept", func() {
			concepts, _ := s.ListConcepts()
			So(len(concepts), ShouldBeGreaterThan, 0)
		})

		Convey("Concept should be findable by name", func() {
			savedConcept, _ := s.LoadConcept(conceptId)
			Convey("User should match except for userID", func() {
				So(savedConcept.Name, ShouldEqual, concept.Name)
				So(savedConcept.Summary, ShouldEqual, concept.Summary)
				So(savedConcept.Full, ShouldEqual, concept.Full)
			})

			Convey("Updating the concept", func() {
				savedConcept.Summary = "a different short version"
				conceptId2, _ := s.UpdateConcept(savedConcept)
				Convey("Concept should keep the same ID and content", func() {
					So(conceptId2, ShouldEqual, conceptId)
					reloadedConcept, _ := s.FindConcept(name)
					So(reloadedConcept.ID, ShouldEqual, savedConcept.ID)
					So(reloadedConcept.Name, ShouldEqual, savedConcept.Name)
					So(reloadedConcept.Summary, ShouldEqual, savedConcept.Summary)
					So(reloadedConcept.Full, ShouldEqual, savedConcept.Full)
				})
			})
		})
	})
}

func TestStore_DeleteConcept(t *testing.T) {
	const name = "test"
	Convey("Given that we have saved a user", t, func() {
		s.PurgeConcept(name)
		concept := Concept{Name: name}
		conceptId, _ := s.InsertConcept(&concept)

		Convey("Concept should be given an ID", func() {
			So(conceptId, ShouldBeGreaterThan, 0)
		})

		Convey("Then I delete the concept", func() {
			s.PurgeConcept(name)

			Convey("Concept should not be findable by email address", func() {
				savedUser, err := s.FindConcept(name)

				So(err, ShouldNotBeNil)
				So(savedUser, ShouldEqual, (*Concept)(nil))
			})
		})
	})
}

func TestStore_AddTagsToConcept(t *testing.T) {
	const name = "test"
	const tag1 = "tag1"
	const tag2 = "tag2"
	Convey("Insert a concept to the store", t, func() {
		s.PurgeConcept(name)
		concept := Concept{Name: name}
		concept.Summary = "a short version of the test concept"
		conceptId, _ := s.InsertConcept(&concept)

		Convey("Concept should be given an ID", func() {
			So(conceptId, ShouldBeGreaterThan, 0)
		})

		Convey("Add Tag", func() {
			conceptTag1 := ConceptTag{Tag: tag1, ConceptId: conceptId, Order: 0}
			conceptTag2 := ConceptTag{Tag: tag2, ConceptId: conceptId, Order: 1}
			conceptTag1Id, _ := s.InsertConceptTag(&conceptTag1)
			conceptTag2Id, _ := s.InsertConceptTag(&conceptTag2)

			Convey("conceptTags should be given an ID", func() {
				So(conceptTag1Id, ShouldBeGreaterThan, 0)
				So(conceptTag2Id, ShouldBeGreaterThan, 0)
			})

			Convey("Concept tags list should contain both names", func() {
				tags, _ := s.ConceptTagsAsStrings(&concept)
				So(len(tags), ShouldEqual, 2)
				So(tags[0], ShouldEqual, tag1)
				So(tags[1], ShouldEqual, tag2)
			})
		})
	})
}

func ensureTestConceptExists(name string) *Concept {
	concept, err := s.FindConcept(name)
	if err != nil {
		concept = &Concept{
			Name:    name,
			Summary: "a short version of the test concept",
		}
		_, _ = s.InsertConcept(concept)
	}
	return concept
}

func TestStore_ListAllTags(t *testing.T) {
	const tagA = "tagA"
	const tagB = "tagB"
	const tagC = "tagC"
	s.PurgeConceptTag(tagA)
	s.PurgeConceptTag(tagB)
	s.PurgeConceptTag(tagC)
	concept := ensureTestConceptExists("testConcept")
	Convey("Create some tags", t, func() {
		conceptTagA := ConceptTag{Tag: tagA, ConceptId: concept.ID, Order: 0}
		conceptTagB := ConceptTag{Tag: tagB, ConceptId: concept.ID, Order: 1}
		conceptTagC := ConceptTag{Tag: tagC, ConceptId: concept.ID, Order: 1}
		conceptTagAId, _ := s.InsertConceptTag(&conceptTagA)
		conceptTagBId, _ := s.InsertConceptTag(&conceptTagB)
		conceptTagCId, _ := s.InsertConceptTag(&conceptTagC)
		Convey("All tags list should contain items", func() {
			tags, _ := s.ListConceptTags()
			tagAFromTags := getTagFromTags(tags, tagA)
			So(tagAFromTags.Tag, ShouldEqual, tagA)
			So(tagAFromTags.ID, ShouldEqual, conceptTagAId)
			tagBFromTags := getTagFromTags(tags, tagB)
			So(tagBFromTags.Tag, ShouldEqual, tagB)
			So(tagBFromTags.ID, ShouldEqual, conceptTagBId)
			tagCFromTags := getTagFromTags(tags, tagC)
			So(tagCFromTags.Tag, ShouldEqual, tagC)
			So(tagCFromTags.ID, ShouldEqual, conceptTagCId)
		})
	})
}

func getTagFromTags(tags []ConceptTag, tag string) *ConceptTag {
	for _, conceptTag := range tags {
		if conceptTag.Tag == tag {
			return &conceptTag
		}
	}
	return nil
}

func getItemFromItems(items []Item, id uint) *Item {
	for _, item := range items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}
func getItemFromItemsOS(items []ItemWithOS, id uint) *ItemWithOS {
	for _, item := range items {
		if item.ID == id {
			return &item
		}
	}
	return nil
}

func TestStore_ItemCreation(t *testing.T) {
	Convey("Create a item", t, func() {
		user1 := ensureTestUserExists("user1@example.com", UserPermissionsUser)
		s.db.Unscoped().Where("user_id=?", user1.ID).Delete(Item{})
		item := Item{
			ViewPermissions:  UserPermissionsUser,
			SilverNumber:     007,
			UserId:           user1.ID,
			Name:             "Oak",
			Description:      "",
			LatitudeI:        12,
			LongitudeI:       34,
			Altitude:         119,
			Status:           "",
			AgeGroup:         "",
			Height:           5.5,
			LatinName:        "",
			DiameterAt:       1.2,
			RenderHints:      "",
			FruitType:        "",
			CroppingSeason:   "",
			FruitStorage:     "",
			PollinatingGroup: "",
			Rootstock:        "",
		}
		itemId, _ := s.InsertItem(&item)
		Convey("Item should be created", func() {
			itemsBy, _ := s.ListItemsByUser(user1.ID, 0, 200)
			itemByFromItems := getItemFromItems(itemsBy, item.ID)
			So(itemByFromItems.ID, ShouldEqual, itemId)

			items, _ := s.ListItemsForUser(user1.ID, 0, 200)
			itemFromItems := getItemFromItems(items, item.ID)
			So(itemFromItems.ID, ShouldEqual, itemId)

			Convey("Updating the item", func() {
				itemFromItems.Status = "died"
				itemId2, _ := s.UpdateItem(itemFromItems)
				Convey("Item should keep the same ID and content", func() {
					So(itemId2, ShouldEqual, itemId)
					reloadedItem, _ := s.LoadItem(itemId2)
					So(reloadedItem.ID, ShouldEqual, itemFromItems.ID)
					So(reloadedItem.Name, ShouldEqual, itemFromItems.Name)
					So(reloadedItem.LatitudeI, ShouldEqual, itemFromItems.LatitudeI)
					So(reloadedItem.LongitudeI, ShouldEqual, itemFromItems.LongitudeI)
					So(reloadedItem.Height, ShouldEqual, itemFromItems.Height)
					So(reloadedItem.Status, ShouldEqual, itemFromItems.Status)
				})
			})
		})
	})
}

func TestStore_ItemCreationWithOS(t *testing.T) {
	Convey("Create a item", t, func() {
		user1 := ensureTestUserExists("user1@example.com", UserPermissionsUser)
		s.db.Unscoped().Where("user_id=?", user1.ID).Delete(Item{})
		item := Item{
			ViewPermissions: UserPermissionsUser,
			SilverNumber:    007,
			UserId:          user1.ID,
			Name:            "TestObject",
			LatitudeI:       559036000,
			LongitudeI:      -31550000,
		}
		itemId, _ := s.InsertItem(&item)
		Convey("Item should be created and readable by region with osgb coordinates", func() {
			itemsFor, err := s.ListItemsForUserWithRegion(user1.ID, 3, 0, 200)
			So(err, ShouldBeNil)
			log.Print(item.ID)
			log.Print(itemsFor)
			itemByFromItems := getItemFromItemsOS(itemsFor, item.ID)
			So(itemByFromItems.ID, ShouldEqual, itemId)
			So(itemByFromItems.OSGB.Lat, ShouldEqual, 668434.8010845537) //Northing
			So(itemByFromItems.OSGB.Lng, ShouldEqual, 327884.0766195121) //Easting
		})
	})
}

func TestStore_TestDays(t *testing.T) {
	Convey("Create a few dates", t, func() {
		date := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		date1 := PosixDateTime(date)
		date2 := PosixDateTime(date.Add(time.Hour * 24 * 1))
		date3 := PosixDateTime(date.Add(time.Hour * 24 * 31))
		date4 := PosixDateTime(date.Add(time.Hour * 24 * 30 * 12))
		So(date1.Days(), ShouldEqual, 20454)
		So(date2.Days(), ShouldEqual, 20455)
		So(date3.Days(), ShouldEqual, 20485)
		So(date4.Days(), ShouldEqual, 20814)
	})
}

func getItemLogFromItemLogs(itemLogs []ItemLog, id uint) *ItemLog {
	for _, itemLog := range itemLogs {
		if itemLog.ID == id {
			return &itemLog
		}
	}
	return nil
}

func TestStore_ItemLogCreation(t *testing.T) {
	Convey("Create a item log", t, func() {
		user1 := ensureTestUserExists("user1@example.com", UserPermissionsEditor)
		s.db.Unscoped().Where("user_id=?", user1.ID).Delete(Item{})
		s.db.Unscoped().Where("unique_id=?", "2").Delete(ItemLog{})
		item := Item{
			SilverNumber: 007,
			UserId:       user1.ID,
			Name:         "Oak",
			LatitudeI:    12,
			LongitudeI:   34,
			Altitude:     119,
		}
		itemId, _ := s.InsertItem(&item)
		itemLog := ItemLog{
			UserId:   user1.ID,
			UniqueId: "2",
			ItemId:   itemId,
			Name:     "Booted",
		}
		itemLogId, _ := s.InsertItemLog(&itemLog)
		Convey("ItemLog should be created", func() {
			itemLogs, _ := s.ListItemLogs(user1.ID, 0, 200)
			itemLogFromItemLogs := getItemLogFromItemLogs(itemLogs, itemLog.ID)
			So(itemLogFromItemLogs.ID, ShouldEqual, itemLogId)
		})
	})
}
