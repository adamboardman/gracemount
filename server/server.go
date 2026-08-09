package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/adamboardman/gracemount/binary"
	"github.com/adamboardman/gracemount/store"
	"github.com/adamboardman/gracemount/tag_updater"
	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type WebApp struct {
	Router                    *gin.Engine
	Store                     *store.Store
	JwtAuthRequiredMiddleware *jwt.GinJWTMiddleware
	JwtAuthAvailMiddleware    *jwt.GinJWTMiddleware
}

var App *WebApp

func (a *WebApp) Init() {
	App = a
	a.Store = &store.Store{}
	a.Store.StoreInit()

	// Set the router as the default one shipped with Gin
	router := gin.Default()
	a.Router = router

	addWebAppStaticFiles(router)
	addApiRoutes(a, router)
	//addPhotoRoutes(a, router)
	addDefaultRouteToWebApp(router)

	performAnyRemainingDailyTasks()
	addTickerForDailyTasks()
}

func addTickerForDailyTasks() {
	ticker := time.NewTicker(12 * time.Hour)
	go func() {
		for {
			fmt.Println("Performing for any daily outstanding tasks")
			performAnyRemainingDailyTasks()
			<-ticker.C
		}
	}()
}

func performAnyRemainingDailyTasks() {
	//TODO - Update statistics?
}

func addApiRoutes(a *WebApp, router *gin.Engine) {
	api := router.Group("/api")
	api.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "root of the API does nothing, next?"})
	})

	a.JwtAuthRequiredMiddleware = a.InitAuth(api)
	api.GET("/concepts", ConceptsList)
	api.GET("/concepts/:conceptID", LoadConcept)
	api.GET("/concepts/:conceptID/tags", LoadConceptTags)
	api.GET("/concept/:tag", FetchConcept)
	api.GET("/concept_tags", ConceptTagsList)
	api.POST("/message", ReceiveMessage)
	api.GET("/regions", RegionsList)
	a.JwtAuthAvailMiddleware = a.AuthAvailableMiddleware()
	api.Use(a.JwtAuthAvailMiddleware.MiddlewareFunc())
	{
		api.GET("/items", ItemsList)
		api.GET("/items/:itemID", LoadItem)
		api.GET("/items/:itemID/logs", LoadItemLogs)
		api.GET("/items_for_region/:regionID", ItemsListForRegion)
	}
	api.Use(a.JwtAuthRequiredMiddleware.MiddlewareFunc())
	{
		api.GET("/users", LoadUsers)
		api.GET("/users/:userID", LoadUser)
		//api.GET("/users/:userID/photo",  UserPhoto)
		//api.POST("/users/:userID/photo", AddUserPhoto)
		//api.PUT("/users/:userID/photo",  UpdateUserPhoto)
		api.PUT("/users/:userID", UpdateUser)
		api.POST("/concepts", EditorPermissionsRequired(), AddConcept)
		api.PUT("/concepts/:conceptID", EditorPermissionsRequired(), UpdateConcept)
		api.POST("/concept_tags", EditorPermissionsRequired(), AddConceptTag)
		api.DELETE("/concept_tags/:conceptTagID", EditorPermissionsRequired(), DeleteConceptTag)
		api.DELETE("/concept_tags", EditorPermissionsRequired(), DeleteConceptTags)
		api.DELETE("/items/:itemID", EditorPermissionsRequired(), DeleteItem)
		api.POST("/items", AddItem)
		api.PUT("/items/:itemID", UpdateItem)
		api.POST("/item_logs", AddItemLog)
		api.GET("/item_logs", ItemLogsList)
	}
}

func EditorPermissionsRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		EditorPermissionsRequiredImpl(c)
	}
}

func EditorPermissionsRequiredImpl(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userId := uint(claims[identityId].(float64))
	user, err := App.Store.LoadPrivilegedUser(userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
		return
	}
	if !(user.Permissions >= store.UserPermissionsEditor) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"statusText": "User is not an editor"})
		return
	}
	c.Next()
}

func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func addDefaultRouteToWebApp(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		if Exists("./public/index.html") {
			c.File("./public/index.html")
		} else {
			c.File("../public/index.html")
		}
	})
}

func (a *WebApp) Run(addr string) {
	_ = a.Router.Run(addr)
}

func addWebAppStaticFiles(router *gin.Engine) {
	router.Static("/public", "./public")
	router.Use(static.Serve("/dist", static.LocalFile("./dist", true)))
}

func LoadUsers(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	user, err := App.Store.LoadPrivilegedUser(loggedInUserId)
	if err != nil || user.Permissions != store.UserPermissionsAdmin {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid UserID"})
		return
	}

	c.Header("Content-Type", "application/json")
	users, err := App.Store.ListUsers()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Users not listable"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func LoadUser(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	c.Header("Content-Type", "application/json")
	user, err := App.Store.LoadPrivilegedUser(loggedInUserId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
		return
	}
	userId, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid UserID"})
		return
	}
	if userId == 0 || uint(userId) == loggedInUserId {
		c.JSON(http.StatusOK, user)
	} else if user.Permissions == store.UserPermissionsAdmin {
		user, err = App.Store.LoadPrivilegedUser(uint(userId))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func UpdateUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("UserID - err: %s", err.Error())})
		return
	}

	user, err := readJSONIntoUser(uint(userId), c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("User details failed validation - err: %s", err.Error())})
		return
	}

	_, err = App.Store.UpdateUser(user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("User details failed validation - err: %s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK, "message": "User updated successfully", "resourceId": userId,
	})
}

func readJSONIntoUser(userId uint, c *gin.Context) (*store.User, error) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	loggedInUser, err := App.Store.LoadPrivilegedUser(loggedInUserId)
	if err != nil {
		return nil, err
	}
	var user *store.User
	if loggedInUser.Permissions != store.UserPermissionsAdmin {
		if userId != loggedInUserId {
			err := errors.New("Only the logged in user can update their profile")
			return nil, err
		}
		user, err = App.Store.LoadUser(loggedInUserId)
		if err != nil {
			return nil, err
		}
	} else {
		user, err = App.Store.LoadUser(userId)
		if err != nil {
			return nil, err
		}
	}
	userJson := UserJSON{}
	err = c.BindJSON(&userJson)
	if err != nil {
		return nil, err
	}

	user.FirstName = userJson.FirstName
	user.MidNames = userJson.MidNames
	user.LastName = userJson.LastName
	user.Location = userJson.Location
	user.PhotoId = userJson.PhotoId
	user.Email = userJson.Email
	user.Mobile = userJson.Mobile
	if loggedInUser.Permissions == store.UserPermissionsAdmin {
		user.Permissions = store.UserPermissions(userJson.Permissions)
	}

	return user, err
}

type UserJSON struct {
	FirstName   string `form:"FirstName" binding:"required"`
	MidNames    string `form:"MidNames" binding:"-"`
	LastName    string `form:"LastName" binding:"-"`
	Location    string `form:"Location" binding:"-"`
	PhotoId     uint   `form:"PhotoId" binding:"-"`
	Email       string `form:"Email" binding:"required"`
	Mobile      string `form:"Mobile" binding:"-"`
	Permissions uint   `form:"Permissions" binding:"-"`
}

func ConceptsList(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	concepts, err := App.Store.ListConcepts()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": fmt.Sprintf("Concepts not found")})
	} else {
		c.JSON(http.StatusOK, concepts)
	}
}

func RegionsList(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	regions, err := App.Store.ListRegions()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": fmt.Sprintf("Regions not found")})
	} else {
		c.JSON(http.StatusOK, regions)
	}
}

func LoadConcept(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	conceptId, err := strconv.Atoi(c.Param("conceptID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid ConceptID"})
		return
	}
	concept, err := App.Store.LoadConcept(uint(conceptId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Concept not found"})
		return
	}
	conceptJSON := ConceptJSON{}
	conceptJSON.ID = concept.ID
	conceptJSON.Name = concept.Name
	conceptJSON.Summary = concept.Summary
	conceptJSON.Full = concept.Full
	c.JSON(http.StatusOK, conceptJSON)
}

func FetchConcept(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	tag := c.Param("tag")
	conceptTag, err := App.Store.FindConceptTag(tag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Concept Tag not found"})
		return
	}
	concept, err := App.Store.LoadConcept(conceptTag.ConceptId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Concept for Tag not found"})
		return
	}
	conceptJSON := ConceptJSON{}
	conceptJSON.ID = concept.ID
	conceptJSON.Name = concept.Name
	conceptJSON.Summary = concept.Summary
	conceptJSON.Full = concept.Full
	c.JSON(http.StatusOK, conceptJSON)
}

func AddConcept(c *gin.Context) {
	concept := store.Concept{}

	err := readJSONIntoConcept(&concept, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Concept failed validation - err: %s", err.Error())})
		return
	}

	conceptTags, _ := App.Store.ListConceptTags()
	concepts, _ := App.Store.ListConcepts()
	concept.Full = tag_updater.UpdateTags(conceptTags, concepts, concept.Full, concept.ID)

	conceptId, err := App.Store.InsertConcept(&concept)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Insert Concept failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated, "message": "Concept created successfully", "resourceId": conceptId,
	})
}

func UpdateConcept(c *gin.Context) {
	conceptId, err := strconv.Atoi(c.Param("conceptID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("ConceptID invalid - err: %s", err.Error())})
		return
	}

	concept := &store.Concept{}
	concept, err = App.Store.LoadConcept(uint(conceptId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Concept not found"})
		return
	}

	err = readJSONIntoConcept(concept, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Concept details failed validation - err: %s", err.Error())})
		return
	}

	conceptTags, _ := App.Store.ListConceptTags()
	concepts, _ := App.Store.ListConcepts()
	concept.Full = tag_updater.UpdateTags(conceptTags, concepts, concept.Full, concept.ID)

	_, err = App.Store.UpdateConcept(concept)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK, "message": "Concept updated successfully", "resourceId": conceptId,
		})
	}
}

func readJSONIntoConcept(concept *store.Concept, c *gin.Context, forceUpdate bool) error {
	conceptJSON := ConceptJSON{}
	err := c.ShouldBindJSON(&conceptJSON)
	if err != nil {
		return err
	}

	if forceUpdate || conceptJSON.ID == 0 {
		concept.ID = conceptJSON.ID
		concept.Name = conceptJSON.Name
		concept.Summary = conceptJSON.Summary
		concept.Full = conceptJSON.Full
	}
	return nil
}

type ConceptJSON struct {
	ID      uint
	Name    string
	Summary string
	Full    string
}

func ConceptTagsList(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	tags, err := App.Store.ListConceptTags()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "ConceptTags not found"})
	} else {
		c.JSON(http.StatusOK, tags)
	}
}

type ConceptTagJSON struct {
	ID        uint
	Tag       string
	ConceptId uint
}

func readJSONIntoConceptTag(conceptTag *store.ConceptTag, c *gin.Context, forceUpdate bool) error {
	conceptTagJSON := ConceptTagJSON{}
	err := c.ShouldBindJSON(&conceptTagJSON)
	if err != nil {
		return err
	}

	if forceUpdate || conceptTagJSON.ID == 0 {
		conceptTag.ID = conceptTagJSON.ID
		conceptTag.Tag = conceptTagJSON.Tag
		conceptTag.ConceptId = conceptTagJSON.ConceptId
	}
	return nil
}

func AddConceptTag(c *gin.Context) {
	conceptTag := store.ConceptTag{}

	err := readJSONIntoConceptTag(&conceptTag, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Concept failed validation - err: %s", err.Error())})
		return
	}

	conceptTagId, err := App.Store.InsertConceptTag(&conceptTag)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Insert Concept Tag failed - err: %s", err.Error())})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated, "message": "Concept Tag created successfully", "resourceId": conceptTagId,
	})
}

func DeleteConceptTag(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	id, err := strconv.Atoi(c.Param("conceptTagID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Invalid ConceptTagID - err: %s", err.Error())})
		return
	}
	err = App.Store.DeleteConceptTag(uint(id))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Delete ConceptTag Failed - err: %s", err.Error())})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK, "message": "ConceptTag deleted", "resourceId": id,
		})
	}
}

type ConceptTagIDs []uint

func DeleteConceptTags(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	conceptTagIDs := ConceptTagIDs{}
	err := c.BindJSON(&conceptTagIDs)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Invalid ConceptTagIDs - err: %s", err.Error())})
		return
	}
	for _, id := range conceptTagIDs {
		err = App.Store.DeleteConceptTag(uint(id))
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Delete ConceptTag Failed - err: %s", err.Error())})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK, "message": "ConceptTag deleted", "resourceIds": conceptTagIDs,
		})
	}
}

func LoadConceptTags(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	conceptId, err := strconv.Atoi(c.Param("conceptID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid ConceptId"})
		return
	}
	conceptTags, err := App.Store.ConceptTagsForConceptId(uint(conceptId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "ConceptTags for Concept not found"})
		return
	}
	c.JSON(http.StatusOK, conceptTags)
}

type ItemJSON struct {
	ID               uint
	ViewPermissions  uint
	UserId           uint
	SilverNumber     uint64
	Name             string
	ItemType         uint
	Description      string
	LatitudeI        int32
	LongitudeI       int32
	Altitude         float32
	Status           string
	AgeGroup         string
	Height           float32
	LatinName        string
	DiameterAt       float32
	RenderHints      string
	FruitType        string
	CroppingSeason   string
	FruitStorage     string
	PollinatingGroup string
	Rootstock        string
}

func readJSONIntoItem(item *store.Item, c *gin.Context, forceUpdate bool) error {
	itemJSON := ItemJSON{}
	err := c.BindJSON(&itemJSON)
	if err != nil {
		return err
	}

	if forceUpdate || itemJSON.ID == 0 {
		item.ID = itemJSON.ID
		item.ViewPermissions = store.UserPermissions(itemJSON.ViewPermissions)
		item.UserId = itemJSON.UserId
		item.SilverNumber = itemJSON.SilverNumber
		item.Name = itemJSON.Name
		item.ItemType = itemJSON.ItemType
		item.Description = itemJSON.Description
		item.LatitudeI = itemJSON.LatitudeI
		item.LongitudeI = itemJSON.LongitudeI
		item.Altitude = itemJSON.Altitude
		item.Height = itemJSON.Height
		item.Status = itemJSON.Status
		item.AgeGroup = itemJSON.AgeGroup
		item.LatinName = itemJSON.LatinName
		item.DiameterAt = itemJSON.DiameterAt
		item.RenderHints = itemJSON.RenderHints
		item.FruitType = itemJSON.FruitType
		item.CroppingSeason = itemJSON.CroppingSeason
		item.FruitStorage = itemJSON.FruitStorage
		item.PollinatingGroup = itemJSON.PollinatingGroup
		item.Rootstock = itemJSON.Rootstock
	}

	return nil
}

func AddItem(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	item := store.Item{}

	err := readJSONIntoItem(&item, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Item failed validation - error: %s", err.Error())})
		return
	}

	if item.ID > 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Trying to add a new item whilst specifying an id"})
		return
	}

	item.UserId = loggedInUserId

	itemId, err := App.Store.InsertItem(&item)
	if err != nil || itemId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Insert Item failed - error: %s", err.Error())})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated, "message": "Item created successfully", "resourceId": itemId,
	})
}

func LoadItem(c *gin.Context) {
	loggedInUserId := uint(0)
	claims := jwt.ExtractClaims(c)
	if claims != nil && claims["id"] != nil {
		loggedInUserId = uint(claims["id"].(float64))
	}

	c.Header("Content-Type", "application/json")
	itemId, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid ItemID"})
		return
	}
	item, err := App.Store.LoadItem(uint(itemId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Item not found"})
		return
	}

	if item.UserId != loggedInUserId {
		user, err := App.Store.LoadPrivilegedUser(loggedInUserId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
			return
		}
		if !(user.Permissions >= store.UserPermissionsEditor || user.Permissions >= item.ViewPermissions) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Item not found"})
			return
		}
	}

	itemJSON := ItemJSON{}
	itemJSON.ID = item.ID
	itemJSON.ViewPermissions = uint(item.ViewPermissions)
	itemJSON.UserId = item.UserId
	itemJSON.SilverNumber = item.SilverNumber
	itemJSON.Name = item.Name
	itemJSON.ItemType = item.ItemType
	itemJSON.Description = item.Description
	itemJSON.LatitudeI = item.LatitudeI
	itemJSON.LongitudeI = item.LongitudeI
	itemJSON.Altitude = item.Altitude
	itemJSON.Height = item.Height
	itemJSON.Status = item.Status
	itemJSON.AgeGroup = item.AgeGroup
	itemJSON.Height = item.Height
	itemJSON.LatinName = item.LatinName
	itemJSON.DiameterAt = item.DiameterAt
	itemJSON.RenderHints = item.RenderHints
	itemJSON.FruitType = item.FruitType
	itemJSON.CroppingSeason = item.CroppingSeason
	itemJSON.FruitStorage = item.FruitStorage
	itemJSON.PollinatingGroup = item.PollinatingGroup
	itemJSON.Rootstock = item.Rootstock
	c.JSON(http.StatusOK, itemJSON)
}

func ItemsList(c *gin.Context) {
	loggedInUserId := uint(0)
	claims := jwt.ExtractClaims(c)
	if claims != nil && claims["id"] != nil {
		loggedInUserId = uint(claims["id"].(float64))
	}

	c.Header("Content-Type", "application/json")
	items, err := App.Store.ListItemsForUser(loggedInUserId, 0, 400)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Items not found"})
	} else {
		c.JSON(http.StatusOK, items)
	}
}

func ItemsListForRegion(c *gin.Context) {
	loggedInUserId := uint(0)
	claims := jwt.ExtractClaims(c)
	if claims != nil && claims["id"] != nil {
		loggedInUserId = uint(claims["id"].(float64))
	}

	regionId, err := strconv.Atoi(c.Param("regionID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("RegionID invalid - err: %s", err.Error())})
		return
	}

	c.Header("Content-Type", "application/json")
	if regionId > 0 {
		items, err := App.Store.ListItemsForUserWithRegion(loggedInUserId, uint(regionId), 0, 400)
		if err == nil {
			c.JSON(http.StatusOK, items)
			return
		}
	} else {
		items, err := App.Store.ListItemsForUser(loggedInUserId, 0, 400)
		if err == nil {
			c.JSON(http.StatusOK, items)
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Items not found"})
}

func UpdateItem(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	itemId, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("ItemID invalid - err: %s", err.Error())})
		return
	}

	item := &store.Item{}
	item, err = App.Store.LoadItem(uint(itemId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Item not found"})
		return
	}

	if item.UserId != loggedInUserId {
		user, err := App.Store.LoadPrivilegedUser(loggedInUserId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
			return
		}
		if !(user.Permissions >= store.UserPermissionsEditor) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Item not found"})
			return
		}
	}

	err = readJSONIntoItem(item, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Item details failed validation - err: %s", err.Error())})
		return
	}

	_, err = App.Store.UpdateItem(item)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK, "message": "Item updated successfully", "resourceId": itemId,
		})
	}
}

func DeleteItem(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	id, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Invalid ItemId - err: %s", err.Error())})
		return
	}
	err = App.Store.PurgeItem(uint(id))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("Delete Item Failed - err: %s", err.Error())})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK, "message": "Item deleted", "resourceId": id,
		})
	}
}

type ItemLogJSON struct {
	ID          uint
	UserId      uint
	ItemId      uint
	Name        string
	Description string
}

func readJSONIntoItemLog(itemLog *store.ItemLog, c *gin.Context, forceUpdate bool) error {
	itemLogJSON := ItemLogJSON{}
	err := c.BindJSON(&itemLogJSON)
	if err != nil {
		return err
	}

	if forceUpdate || itemLogJSON.ID == 0 {
		itemLog.ID = itemLogJSON.ID
		itemLog.UserId = itemLogJSON.UserId
		itemLog.ItemId = itemLogJSON.ItemId
		itemLog.Name = itemLogJSON.Name
		itemLog.Description = itemLogJSON.Description
	}

	return nil
}

func AddItemLog(c *gin.Context) {
	itemLog := store.ItemLog{}

	err := readJSONIntoItemLog(&itemLog, c, true)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": fmt.Sprintf("ItemLog failed validation - error: %s", err.Error())})
		return
	}

	itemLogId, err := App.Store.InsertItemLog(&itemLog)
	if err != nil || itemLogId == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Insert ItemLog failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"status": http.StatusCreated, "message": "ItemLog created successfully", "resourceId": itemLogId,
	})
}

func ItemLogsList(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	c.Header("Content-Type", "application/json")
	itemLogs, err := App.Store.ListItemLogs(loggedInUserId, 0, 200)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "ItemLogs not found"})
	} else {
		c.JSON(http.StatusOK, itemLogs)
	}
}

func LoadItemLogs(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	loggedInUserId := uint(claims["id"].(float64))

	c.Header("Content-Type", "application/json")

	itemId, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Invalid ItemId"})
		return
	}

	item := &store.Item{}
	item, err = App.Store.LoadItem(uint(itemId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Item not found"})
		return
	}

	if item.UserId != loggedInUserId {
		user, err := App.Store.LoadPrivilegedUser(loggedInUserId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "User not found"})
			return
		}
		if !(user.Permissions >= store.UserPermissionsEditor || user.Permissions >= item.ViewPermissions) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "ItemLogs not found"})
			return
		}
	}

	itemLogs, err := App.Store.ListItemLogsForItem(uint(itemId))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "ItemLogs not found"})
	} else {
		c.JSON(http.StatusOK, itemLogs)
	}
}

func ReceiveMessage(c *gin.Context) {
	/**
	  No login required
	   - decode message / decrypt if encrypted for us
	   - create item (with UserPermissionsEditor if not found)
	   - add ItemLog to Item
	*/
	data, err := c.GetRawData()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Invalid Message Format"})
		return
	}
	reader := binary.BinaryReader{
		Offset:     0,
		Buffer:     data,
		BufferSize: uint32(len(data)),
		Pos:        0,
	}
	var message binary.Message
	message.ReadIn(&reader)

	c.Header("Content-Type", "application/json")
	if message.Malformed || message.Channel != "#smoke" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Invalid Message Format"})
	} else {
		item := &store.Item{}
		item, err = App.Store.LoadItemBySilverNumber(message.PacketSenderId)
		if err != nil {
			item.SilverNumber = message.PacketSenderId
		}
		item.ViewPermissions = store.UserPermissionsEditor
		if message.SenderNickname != "anon" {
			item.Name = message.SenderNickname
		}
		item.LatitudeI = message.LatitudeI
		item.LongitudeI = message.LongitudeI
		item.Altitude = float32(message.Altitude)
		var itemId uint
		if err != nil {
			itemId, err = App.Store.InsertItem(item)
		} else {
			itemId, err = App.Store.UpdateItem(item)
		}

		itemLog, err := App.Store.LoadItemLogByUniqueId(message.MessageId)
		if err == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"statusText": "Duplicate Message"})
			return
		}
		itemLog.ItemId = itemId
		itemLog.CreatedAt = time.UnixMilli(message.PacketTimestampMs)
		itemLog.UniqueId = message.MessageId
		itemLog.Name = message.Content

		itemLogId, err := App.Store.InsertItemLog(itemLog)
		if err != nil || itemLogId == 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"statusText": "Insert ItemLog failed"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"status": http.StatusCreated, "message": "ItemLog created successfully", "resourceId": itemLogId,
		})
	}
}
