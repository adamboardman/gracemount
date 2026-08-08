package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/adamboardman/gracemount/store"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/argon2"
)

type Response struct {
	Token string `json:"token"`
}

var a WebApp

func TestMain(m *testing.M) {
	a = WebApp{}
	a.Init()

	code := m.Run()

	os.Exit(code)
}

func TestWebApp404(t *testing.T) {
	Convey("Open an invalid URL", t, func() {
		req, _ := http.NewRequest("GET", "/invalidurl", nil)
		response := httptest.NewRecorder()
		a.Router.ServeHTTP(response, req)

		Convey("The server should respond with StatusOK (defaults to react web app)", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestWebAppIndex(t *testing.T) {
	Convey("Web Index", t, func() {
		req, _ := http.NewRequest("GET", "/", nil)
		response := httptest.NewRecorder()
		a.Router.ServeHTTP(response, req)

		Convey("The server should respond with StatusOK", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
		})
	})
}

func TestRegisterUser(t *testing.T) {
	Convey("Given test user is not in database", t, func() {
		const emailAddress = "test@example.com"
		a.Store.PurgeUser(emailAddress)

		Convey("The user registers", func() {
			registerJSON := RegisterJSON{}
			registerJSON.Email = emailAddress
			registerJSON.Password = "1234"
			registerJSON.PasswordConfirmation = "1234"
			data, _ := json.Marshal(registerJSON)
			postData := bytes.NewReader(data)
			req, _ := http.NewRequest("POST", "/api/auth/register", postData)
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			a.Router.ServeHTTP(response, req)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			Convey("Should send an email with verification code", func() {
				savedUser, _ := a.Store.FindUser(emailAddress)
				So(savedUser.ConfirmVerifier, ShouldNotBeNil)
			})
		})
	})
}

//func TestInvalidTokenRejection(t *testing.T) {
//	Convey("Refreshing an invalid token", t, func() {
//		token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3Q3QGV4YW1wbGUuY29tIiwiZXhwIjoxNTM2Njc3NzUxLCJvcmlnX2lhdCI6MTUzNjY3NDE1MX0.65PStZIR8yRhJo7w2cF8VL-dtF1CbrOnvdB6ub9GxdY"
//		req, _ := http.NewRequest("GET", "/api/auth/refresh_token", nil)
//		req.Header.Set("Authorization", "Bearer "+token)
//		response := httptest.NewRecorder()
//		a.Router.ServeHTTP(response, req)
//		Convey("Should give an error", func() {
//			So(response.Code, ShouldEqual, http.StatusUnauthorized)
//		})
//	})
//}

func TestConfirmEmail(t *testing.T) {
	Convey("Given test user without confirmation is in database", t, func() {
		const emailAddress = "test-confirmed@example.com"
		const verification = "5678"
		a.Store.PurgeUser(emailAddress)

		salt := RandomBytes(16)
		verificationKey := argon2.IDKey([]byte(verification), salt, 1, 64*1024, 4, 32)
		user := store.User{
			Salt:            base64.StdEncoding.EncodeToString(salt),
			ConfirmVerifier: base64.StdEncoding.EncodeToString(verificationKey),
		}
		user.Email = emailAddress
		_, _ = a.Store.InsertUser(&user)

		Convey("The user confirms email address and is redirected to the web app", func() {
			data := url.Values{}
			data.Set("email", emailAddress)
			data.Set("verification", base64.StdEncoding.EncodeToString([]byte(verification)))
			req, _ := http.NewRequest("GET", "/api/auth/confirm_email?"+data.Encode(), nil)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			a.Router.ServeHTTP(response, req)
			So(response.Code, ShouldEqual, http.StatusTemporaryRedirect)
		})
	})
}

func TestMinimalSiteAccessWithoutConfirmedEmail(t *testing.T) {
	Convey("Given unconfirmed test user", t, func() {
		const emailAddress = "test-unconfirmed@example.com"
		user := ensureTestUserExists(emailAddress, store.UserPermissionsNone)
		user.Confirmed = false
		_, _ = a.Store.UpdateUser(user)

		response := loginToUserJSON(emailAddress)

		Convey("Login should succeed", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
		})

		token := userTokenFromLoginResponse(response)

		Convey("Attempt to read user profile", func() {
			req, _ := http.NewRequest("GET", "/api/users/"+uintToString(user.ID), nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response2 := httptest.NewRecorder()
			a.Router.ServeHTTP(response2, req)

			Convey("Should return error", func() {
				So(response2.Code, ShouldEqual, http.StatusForbidden)
			})
		})
	})
}

//func TestRefreshToken(t *testing.T) {
//	Convey("Given confirmed test user", t, func() {
//		const emailAddress = "test@example.com"
//		user := ensureTestUserExists(emailAddress)
//		user.Confirmed = true
//		_, _ = a.Store.UpdateUser(user)
//
//		response := loginToUserJSON(emailAddress)
//
//		Convey("Login should succeed", func() {
//			So(response.Code, ShouldEqual, http.StatusOK)
//		})
//
//		token := userTokenFromLoginResponse(response)
//
//		Convey("Should be able to refresh a token", func() {
//			req2, _ := http.NewRequest("GET", "/api/auth/refresh_token", nil)
//			req2.Header.Set("Authorization", "Bearer "+token)
//			response2 := httptest.NewRecorder()
//			a.Router.ServeHTTP(response2, req2)
//			So(response2.Code, ShouldEqual, http.StatusOK)
//		})
//	})
//}

func userTokenFromLoginResponse(response *httptest.ResponseRecorder) string {
	body, err := io.ReadAll(response.Body)
	responseData := new(Response)
	err = json.Unmarshal(body, responseData)
	So(err, ShouldBeNil)
	token := responseData.Token
	return token
}

func loginToUser(emailAddress string) *httptest.ResponseRecorder {
	data := url.Values{}
	data.Set("email", emailAddress)
	data.Set("password", "1234")
	postData := strings.NewReader(data.Encode())
	req, _ := http.NewRequest("POST", "/api/auth/login", postData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)
	return response
}

type LoginJSON struct {
	Email    string
	Password string
}

func loginToUserJSON(emailAddress string) *httptest.ResponseRecorder {
	loginJSON := LoginJSON{}
	loginJSON.Email = emailAddress
	loginJSON.Password = "1234"
	data, _ := json.Marshal(loginJSON)
	post_data := bytes.NewReader(data)
	req, _ := http.NewRequest("POST", "/api/auth/login", post_data)
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	a.Router.ServeHTTP(response, req)
	return response
}

func ensureTestUserExists(emailAddress string, permissions store.UserPermissions) *store.User {
	user, err := a.Store.FindUser(emailAddress)
	if err != nil {
		salt := RandomBytes(16)
		encrypted := argon2.IDKey([]byte("1234"), salt, 1, 64*1024, 4, 32)
		user = &store.User{
			Salt:     base64.StdEncoding.EncodeToString(salt),
			Password: base64.StdEncoding.EncodeToString(encrypted),
		}
		user.Email = emailAddress
		user.Confirmed = true
		user.Permissions = permissions
		_, _ = a.Store.InsertUser(user)
	}
	return user
}

func TestConceptsList(t *testing.T) {
	Convey("The concepts list should be get'able", t, func() {
		req, _ := http.NewRequest("GET", "/api/concepts", nil)
		response := httptest.NewRecorder()
		a.Router.ServeHTTP(response, req)

		Convey("The server should respond with StatusOK", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
			body, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)
			responseData := new([]store.Concept)
			err = json.Unmarshal(body, responseData)
			So(err, ShouldBeNil)
		})
	})
}

func TestConceptTagsList(t *testing.T) {
	Convey("The concept tags list should be get'able", t, func() {
		req, _ := http.NewRequest("GET", "/api/concept_tags", nil)
		response := httptest.NewRecorder()
		a.Router.ServeHTTP(response, req)

		Convey("The server should respond with StatusOK", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
			body, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)
			responseData := new([]store.ConceptTag)
			err = json.Unmarshal(body, responseData)
			So(err, ShouldBeNil)
		})
	})
}

func TestConceptTagsListForConcept(t *testing.T) {
	concept := ensureTestConceptExists("testConcept")

	Convey("The concept tags list should be get'able", t, func() {
		req, _ := http.NewRequest("GET", "/api/concepts/"+uintToString(concept.ID)+"/tags", nil)
		response := httptest.NewRecorder()
		a.Router.ServeHTTP(response, req)

		Convey("The server should respond with StatusOK", func() {
			So(response.Code, ShouldEqual, http.StatusOK)
			body, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)
			responseData := new([]store.ConceptTag)
			err = json.Unmarshal(body, responseData)
			So(err, ShouldBeNil)
		})
	})
}

func ensureTestConceptExists(name string) *store.Concept {
	concept, err := a.Store.FindConcept(name)
	if err != nil {
		concept = &store.Concept{
			Name:    name,
			Summary: "a short version of the test concept",
		}
		_, _ = a.Store.InsertConcept(concept)
	}
	return concept
}

func TestAddTagToConceptAsUserShouldFail(t *testing.T) {
	concept := ensureTestConceptExists("testConcept")

	Convey("Given a test user", t, func() {
		const emailAddress = "test-login@example.com"
		ensureTestUserExists(emailAddress, store.UserPermissionsUser)

		tagTag := "LETS"
		a.Store.PurgeConceptTag(tagTag)

		Convey("The user logs in", func() {
			response := loginToUserJSON(emailAddress)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Add a concept tag", func() {
				conceptTagJSON := ConceptTagJSON{}
				conceptTagJSON.Tag = tagTag
				conceptTagJSON.ConceptId = concept.ID
				data, _ := json.Marshal(conceptTagJSON)
				post_data := bytes.NewReader(data)
				req2, _ := http.NewRequest("POST", "/api/concept_tags", post_data)
				req2.Header.Set("Content-Type", "application/json")
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with error", func() {
					So(response2.Code, ShouldEqual, http.StatusForbidden)

					tags, err := a.Store.ListConceptTags()
					So(err, ShouldBeNil)

					found := checkArrayForConceptTag(tags, tagTag)
					So(found, ShouldBeFalse)
				})
			})
		})
	})
}

func TestAddTagToConceptAsEditor(t *testing.T) {
	concept := ensureTestConceptExists("testConcept")

	Convey("Given a test editor user", t, func() {
		const emailAddress = "test-editor@example.com"
		user := ensureTestUserExists(emailAddress, store.UserPermissionsEditor)
		So(user, ShouldNotBeNil)

		tagTag := "LETS"
		a.Store.PurgeConceptTag(tagTag)

		Convey("The user logs in", func() {
			response := loginToUserJSON(emailAddress)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Add a concept tag", func() {
				conceptTagJSON := ConceptTagJSON{}
				conceptTagJSON.Tag = tagTag
				conceptTagJSON.ConceptId = concept.ID
				data, _ := json.Marshal(conceptTagJSON)
				post_data := bytes.NewReader(data)
				req2, _ := http.NewRequest("POST", "/api/concept_tags", post_data)
				req2.Header.Set("Content-Type", "application/json")
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with StatusCreated and the tag should be added", func() {
					So(response2.Code, ShouldEqual, http.StatusCreated)

					tags, err := a.Store.ListConceptTags()
					So(err, ShouldBeNil)

					found := checkArrayForConceptTag(tags, tagTag)
					So(found, ShouldBeTrue)
				})
			})
		})
	})
}

func checkArrayForConceptTag(tags []store.ConceptTag, tagTag string) bool {
	found := false
	for _, tag := range tags {
		if tag.Tag == tagTag {
			found = true
		}
	}
	return found
}

func uintToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func TestDeleteTagAsUserFails(t *testing.T) {
	Convey("Given a test user", t, func() {
		const emailAddress = "test-login@example.com"
		ensureTestUserExists(emailAddress, store.UserPermissionsUser)

		tagTag := "LETS"
		a.Store.PurgeConceptTag(tagTag)
		concept := ensureTestConceptExists("testConcept")
		conceptTag := store.ConceptTag{
			Tag:       tagTag,
			ConceptId: concept.ID,
		}
		conceptTagId, err := a.Store.InsertConceptTag(&conceptTag)
		So(err, ShouldBeNil)

		Convey("The user logs in", func() {
			response := loginToUserJSON(emailAddress)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Delete tag", func() {
				req2, _ := http.NewRequest("DELETE", "/api/concept_tags/"+uintToString(conceptTagId), nil)
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with StatusForbidden", func() {
					So(response2.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})
	})
}

func TestDeleteTagAsAdmin(t *testing.T) {
	Convey("Given a test user", t, func() {
		const emailAddress = "test-admin@example.com"
		user := ensureTestUserExists(emailAddress, store.UserPermissionsAdmin)
		So(user, ShouldNotBeNil)
		tagTag := "LETS"
		a.Store.PurgeConceptTag(tagTag)
		concept := ensureTestConceptExists("testConcept")
		conceptTag := store.ConceptTag{
			Tag:       tagTag,
			ConceptId: concept.ID,
		}
		conceptTagId, err := a.Store.InsertConceptTag(&conceptTag)
		So(err, ShouldBeNil)

		Convey("The user logs in", func() {
			response := loginToUserJSON(emailAddress)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Delete tag", func() {
				req2, _ := http.NewRequest("DELETE", "/api/concept_tags/"+uintToString(conceptTagId), nil)
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with StatusOK and the tag should be removed", func() {
					So(response2.Code, ShouldEqual, http.StatusOK)

					conceptTags, err := a.Store.ListConceptTags()
					So(err, ShouldBeNil)

					found := checkArrayForConceptTag(conceptTags, tagTag)
					So(found, ShouldBeFalse)
				})
			})
		})
	})
}

type ApiActionResponse struct {
	Message     string
	ResourceId  uint
	ResourceIds []uint
	Status      uint
}

func TestDeleteTagsAsEditor(t *testing.T) {
	Convey("Given a test user", t, func() {
		const emailAddress = "test-editor@example.com"
		user := ensureTestUserExists(emailAddress, store.UserPermissionsEditor)
		So(user, ShouldNotBeNil)
		concept := ensureTestConceptExists("testConcept")

		tagTag1 := "LETS"
		a.Store.PurgeConceptTag(tagTag1)
		conceptTag1 := store.ConceptTag{
			Tag:       tagTag1,
			ConceptId: concept.ID,
		}
		conceptTag1Id, err := a.Store.InsertConceptTag(&conceptTag1)
		So(err, ShouldBeNil)

		tagTag2 := "local exchange trading system"
		a.Store.PurgeConceptTag(tagTag2)
		conceptTag2 := store.ConceptTag{
			Tag:       tagTag2,
			ConceptId: concept.ID,
		}
		conceptTag2Id, err := a.Store.InsertConceptTag(&conceptTag2)
		So(err, ShouldBeNil)

		Convey("The user logs in", func() {
			response := loginToUserJSON(emailAddress)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Delete tags", func() {
				var tags [2]uint
				tags[0] = conceptTag1Id
				tags[1] = conceptTag2Id
				data, _ := json.Marshal(tags)
				post_data := bytes.NewReader(data)

				req2, _ := http.NewRequest("DELETE", "/api/concept_tags", post_data)
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with StatusOK and the tags should be removed", func() {
					So(response2.Code, ShouldEqual, http.StatusOK)
					body, err := io.ReadAll(response2.Body)
					So(err, ShouldBeNil)
					responseData := new(ApiActionResponse)
					err = json.Unmarshal(body, responseData)
					So(err, ShouldBeNil)

					So(responseData.ResourceIds[0], ShouldEqual, conceptTag1Id)
					So(responseData.ResourceIds[1], ShouldEqual, conceptTag2Id)

					conceptTags, err := a.Store.ListConceptTags()
					So(err, ShouldBeNil)

					found1 := checkArrayForConceptTag(conceptTags, tagTag1)
					So(found1, ShouldBeFalse)
					found2 := checkArrayForConceptTag(conceptTags, tagTag2)
					So(found2, ShouldBeFalse)
				})
			})
		})
	})
}

func checkArrayForItem(items []store.Item, itemJson ItemJSON, ignoreToId bool) bool {
	found := false
	for _, item := range items {
		if item.UserId == itemJson.UserId &&
			item.SilverNumber == itemJson.SilverNumber &&
			item.Name == itemJson.Name &&
			item.Description == itemJson.Description &&
			item.LatitudeI == itemJson.LatitudeI &&
			item.LongitudeI == itemJson.LongitudeI &&
			item.Altitude == itemJson.Altitude &&
			item.Status == itemJson.Status &&
			item.AgeGroup == itemJson.AgeGroup &&
			item.Height == itemJson.Height &&
			item.LatinName == itemJson.LatinName &&
			item.DiameterAt == itemJson.DiameterAt &&
			item.RenderHints == itemJson.RenderHints &&
			item.FruitType == itemJson.FruitType &&
			item.CroppingSeason == itemJson.CroppingSeason &&
			item.FruitStorage == itemJson.FruitStorage &&
			item.PollinatingGroup == itemJson.PollinatingGroup &&
			item.Rootstock == itemJson.Rootstock {

			found = true
		}
	}
	return found
}

func TestCreateItem(t *testing.T) {
	Convey("Given a test user", t, func() {
		const emailAddressOrigin = "test-user1@example.com"
		user1 := ensureTestUserExists(emailAddressOrigin, store.UserPermissionsUser)

		Convey("The user1 logs in", func() {
			response := loginToUserJSON(emailAddressOrigin)

			Convey("The server should respond with StatusOK", func() {
				So(response.Code, ShouldEqual, http.StatusOK)
			})

			token := userTokenFromLoginResponse(response)

			Convey("Create item", func() {
				itemJSON := ItemJSON{}
				itemJSON.UserId = user1.ID
				itemJSON.Name = "Sycamore"
				itemJSON.SilverNumber = 3
				ClearItemsMatchingJSON(itemJSON)
				data, _ := json.Marshal(itemJSON)
				post_data := bytes.NewReader(data)
				req2, _ := http.NewRequest("POST", "/api/items", post_data)
				req2.Header.Set("Content-Type", "application/json")
				req2.Header.Set("Authorization", "Bearer "+token)
				response2 := httptest.NewRecorder()
				a.Router.ServeHTTP(response2, req2)

				Convey("The server should respond with StatusCreated and the item should be created", func() {
					So(response2.Code, ShouldEqual, http.StatusCreated)

					userItems, err := a.Store.ListItemsByUser(user1.ID, 0, 200)
					So(err, ShouldBeNil)

					found := checkArrayForItem(userItems, itemJSON, false)
					So(found, ShouldBeTrue)
				})
			})
		})
	})
}

func ClearItemsMatchingJSON(itemJson ItemJSON) {
	userItems, _ := a.Store.ListItemsForUser(itemJson.UserId, 0, 200)

	for _, item := range userItems {
		if item.SilverNumber == itemJson.SilverNumber {
			a.Store.PurgeItem(item.ID)
		}
	}
}
