package store

import (
	"database/sql/driver"
	"errors"
	"github.com/restayway/gogis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
)

type Store struct {
	db *gorm.DB
}

type PublicUser struct {
	gorm.Model
	Email     string `gorm:"uniqueIndex"`
	FirstName string
	MidNames  string
	LastName  string
	Location  string
	PhotoId   uint
}

func (PublicUser) TableName() string {
	return "users"
}

type UserPermissions uint

const (
	UserPermissionsNone   UserPermissions = 0
	UserPermissionsUser   UserPermissions = 1
	UserPermissionsEditor UserPermissions = 2
	UserPermissionsAdmin  UserPermissions = 3
)

type User struct {
	PrivilegedUser
	Salt               string `json:"-"`
	Password           string `json:"-"`
	ConfirmVerifier    string `json:"-"`
	RecoverVerifier    string `json:"-"`
	RecoverTokenExpiry string `json:"-"`
}

type PrivilegedUser struct {
	PublicUser
	Mobile       string
	Confirmed    bool
	AttemptCount int    `json:"-"`
	LastAttempt  string `json:"-"`
	Locked       string `json:"-"`
	Permissions  UserPermissions
}

type Concept struct {
	gorm.Model
	Name    string
	Summary string
	Full    string
}

type ConceptTag struct {
	gorm.Model
	Tag       string
	ConceptId uint
	Order     uint
}

type Item struct {
	gorm.Model
	UserId           uint
	ViewPermissions  UserPermissions
	SilverNumber     uint64 `gorm:"uniqueIndex"`
	Name             string
	ItemType         uint //ItemType is only used client side so we don't bother replicating the type definitions here
	Description      string
	Location         gogis.Point `gorm:"type:geometry(Point,4326)"`
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

type ItemWithOS struct {
	Item
	OSGB gogis.Point `gorm:"type:geometry(Point,27700)"`
}

type ItemLog struct {
	gorm.Model
	UserId      uint
	ItemId      uint
	UniqueId    string `gorm:"uniqueIndex"`
	Name        string
	Description string
}

type Region struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Name      string        `gorm:"uniqueIndex"`
	Area      gogis.Polygon `gorm:"type:geometry(Polygon,4326)"`
}

type PosixDateTime time.Time

func (d PosixDateTime) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte("0"), nil
	}
	return []byte(strconv.FormatInt(time.Time(d).UTC().UnixMilli(), 10)), nil
}

func (d *PosixDateTime) UnmarshalJSON(b []byte) (err error) {
	p, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return
	}
	t := time.UnixMilli(p).UTC()
	*d = PosixDateTime(t)
	return
}

func (d PosixDateTime) Value() (driver.Value, error) {
	return time.Time(d).UTC(), nil
}

func (d *PosixDateTime) Scan(src interface{}) error {
	if val, ok := src.(time.Time); ok {
		*d = PosixDateTime(val)
	}
	return nil
}

func (d PosixDateTime) AddDate(years int, months int, days int) PosixDateTime {
	return PosixDateTime(time.Time(d).AddDate(years, months, days))
}

func (d PosixDateTime) Days() uint64 {
	unixSeconds := time.Time(d).UTC().Unix()
	secondsPerMinute := 60
	secondsPerHour := 60 * secondsPerMinute
	secondsPerDay := 24 * secondsPerHour
	return uint64(unixSeconds) / uint64(secondsPerDay)
}

func (d PosixDateTime) DateString() string {
	return time.Time(d).Format("2006-01-01")
}

func readPostgresArgs() string {
	const postgresArgsFileName = "postgres_args.txt"
	postgresArgs, err := os.ReadFile(postgresArgsFileName)
	if err != nil {
		postgresArgs, err = os.ReadFile("../" + postgresArgsFileName)
		if err != nil {
			postgresArgs = []byte("host=myhost port=myport sslmode=disable user=myusername dbname=mydbname password=mypassword")
			err = os.WriteFile(postgresArgsFileName, postgresArgs, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return string(postgresArgs)
}

func (s *Store) StoreInit() {
	pg := postgres.Open(readPostgresArgs())
	db, err := gorm.Open(pg, &gorm.Config{
		// DEBUG - add/remove to investigate SQL queries being executed
		// Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal(err)
	}
	s.db = db

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")
	sqlDB.Exec("SET TIME ZONE 'UTC';")

	err = db.AutoMigrate(&User{}, &Concept{}, &ConceptTag{}, &Item{}, &ItemLog{}, &Region{})
	if err != nil {
		log.Fatal(err)
	}

	// We already have these created in our database from the old gorm code where they existed
	// The new model of linking directly to an instance of the Object involves lots of extra
	// database loading so we have rejected upgrading to that, suspect custom Exec lines will
	// be required for any new foreign keys needed.

	// db.Model(&ConceptTag{}).AddForeignKey("concept_id", "concepts(id)", "CASCADE", "RESTRICT")
	// db.Model(&Item{}).AddForeignKey("user_id", "users(id)", "CASCADE", "RESTRICT")
	// db.Model(&ItemLog{}).AddForeignKey("user_id", "users(id)", "CASCADE", "RESTRICT")
	// db.Model(&ItemLog{}).AddForeignKey("item_id", "items(id)", "CASCADE", "RESTRICT")

	s.CreateTESRegion()
	s.CreateGMDTRegion()
	s.CreateMedicalCenterRegion()
}

func (s *Store) CreateRegionOrUpdateRegion(region Region) {
	empty := Region{}
	err := s.db.Where("name=?", region.Name).First(&empty).Error
	if err != nil {
		err = s.db.Create(&region).Error
	}
}

func (s *Store) CreateTESRegion() {
	tes := Region{
		Name: "Gracemount Community Garden (TES)",
		Area: gogis.Polygon{
			Rings: [][]gogis.Point{
				{
					{Lat: 55.9050484, Lng: -3.1562704},
					{Lat: 55.90453, Lng: -3.1562},
					{Lat: 55.90453, Lng: -3.1579337},
					{Lat: 55.9046284, Lng: -3.1579337},
					{Lat: 55.9046662, Lng: -3.1579155},
					{Lat: 55.9047056, Lng: -3.1578869},
					{Lat: 55.9047499, Lng: -3.1578343},
					{Lat: 55.9047905, Lng: -3.1577465},
					{Lat: 55.9048350, Lng: -3.1576178},
					{Lat: 55.9050476, Lng: -3.1569914},
					{Lat: 55.9050553, Lng: -3.1569278},
					{Lat: 55.9050452, Lng: -3.1568443},
					{Lat: 55.9050070, Lng: -3.1567280},
					{Lat: 55.9048259, Lng: -3.1566927},
					{Lat: 55.9050484, Lng: -3.1562704},
				},//55.904520
			},
		},
	}
	s.CreateRegionOrUpdateRegion(tes)
}

func (s *Store) CreateGMDTRegion() {
	gmdt := Region{
		Name: "Gracemount Mansion (GMDT)",
		Area: gogis.Polygon{
			Rings: [][]gogis.Point{
				{
					{Lat: 55.9050484, Lng: -3.1562704},
					{Lat: 55.9048259, Lng: -3.1566927},
					{Lat: 55.9050070, Lng: -3.1567280},
					{Lat: 55.9050452, Lng: -3.1568443},
					{Lat: 55.9050553, Lng: -3.1569278},
					{Lat: 55.9050476, Lng: -3.1569914},
					{Lat: 55.9048350, Lng: -3.1576178},
					{Lat: 55.9047905, Lng: -3.1577465},
					{Lat: 55.9047499, Lng: -3.1578343},
					{Lat: 55.9047056, Lng: -3.1578869},
					{Lat: 55.9046662, Lng: -3.1579155},
					{Lat: 55.9046284, Lng: -3.1579337},
					{Lat: 55.904529, Lng: -3.158003},
					{Lat: 55.90464, Lng: -3.15864},
					{Lat: 55.905671, Lng: -3.15868},
					{Lat: 55.905792, Lng: -3.157111},
					{Lat: 55.9051, Lng: -3.15626},
					{Lat: 55.90506, Lng: -3.156249},
					{Lat: 55.9050484, Lng: -3.1562704},
				},
			},
		},
	}
	s.CreateRegionOrUpdateRegion(gmdt)
}

func (s *Store) CreateMedicalCenterRegion() {
	medical := Region{
		Name: "Gracemount Medical Center Garden (GMG)",
		Area: gogis.Polygon{
			Rings: [][]gogis.Point{
				{
					{Lat: 55.903864, Lng: -3.154628},
					{Lat: 55.903410, Lng: -3.154628},
					{Lat: 55.903410, Lng: -3.155497},
					{Lat: 55.903864, Lng: -3.155497},
					{Lat: 55.903864, Lng: -3.154628},
				},
			},
		},
	}
	s.CreateRegionOrUpdateRegion(medical)
}

func (s *Store) InsertUser(user *User) (uint, error) {
	err := s.db.Create(user).Error
	return user.ID, err
}

func (s *Store) UpdateUser(user *User) (uint, error) {
	err := s.db.Save(user).Error
	return user.ID, err
}

func (s *Store) FindUser(email string) (*User, error) {
	user := User{}
	err := s.db.Where("email=?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (s *Store) PurgeUser(email string) {
	s.db.Unscoped().Where("email=?", email).Delete(User{})
}

func (s *Store) LoadPublicUser(id uint) (*PublicUser, error) {
	user := User{}
	err := s.db.Where("id=?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	publicUser := PublicUser{}
	publicUser.ID = user.ID
	publicUser.FirstName = user.FirstName
	publicUser.MidNames = user.MidNames
	publicUser.LastName = user.LastName
	publicUser.Location = user.Location
	publicUser.PhotoId = user.PhotoId
	return &publicUser, err
}

func (s *Store) LoadPrivilegedUser(userId uint) (*PrivilegedUser, error) {
	user := PrivilegedUser{}
	err := s.db.Where("id=?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (s *Store) LoadUserAsSelf(userId uint, loggedInUserId uint) (*User, error) {
	if userId != loggedInUserId {
		return nil, errors.New("cannot load others users")
	}
	user := User{}
	err := s.db.Where("id=?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (s *Store) LoadUser(userId uint) (*User, error) {
	user := User{}
	err := s.db.Where("id=?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (s *Store) ListUsers() ([]User, error) {
	var users []User
	err := s.db.Limit(200).Order("id").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, err
}

func (s *Store) InsertConcept(concept *Concept) (uint, error) {
	err := s.db.Create(concept).Error
	return concept.ID, err
}

func (s *Store) UpdateConcept(concept *Concept) (uint, error) {
	err := s.db.Save(concept).Error
	return concept.ID, err
}

func (s *Store) PurgeConcept(name string) {
	s.db.Unscoped().Where("name=?", name).Delete(Concept{})
}

func (s *Store) LoadConcept(id uint) (*Concept, error) {
	concept := Concept{}
	err := s.db.Where("id=?", id).First(&concept).Error
	return &concept, err
}

func (s *Store) FindConcept(name string) (*Concept, error) {
	concept := Concept{}
	err := s.db.Where("name=?", name).First(&concept).Error
	if err != nil {
		return nil, err
	}
	return &concept, err
}

func (s *Store) ListConcepts() ([]Concept, error) {
	var concepts []Concept
	err := s.db.Limit(200).Order("name").Find(&concepts).Error
	if err != nil {
		return nil, err
	}
	return concepts, err
}

func (s *Store) ListRegions() ([]Region, error) {
	var regions []Region
	err := s.db.Limit(20).Order("name").Find(&regions).Error
	if err != nil {
		return nil, err
	}
	return regions, err
}

func (s *Store) InsertConceptTag(conceptTag *ConceptTag) (uint, error) {
	err := s.db.Create(conceptTag).Error
	return conceptTag.ID, err
}

func (s *Store) UpdateConceptTag(conceptTag *ConceptTag) (uint, error) {
	err := s.db.Save(conceptTag).Error
	return conceptTag.ID, err
}

func (s *Store) ConceptTagsAsStrings(concept *Concept) ([]string, error) {
	var names []string
	var conceptTags []ConceptTag
	err := s.db.Where("concept_id=?", concept.ID).Order("concept_tags.order").Find(&conceptTags).Error
	if err == nil {
		for _, conceptTag := range conceptTags {
			names = append(names, conceptTag.Tag)
		}
		return names, err
	}
	return nil, err
}

func (s *Store) FindConceptTag(tag string) (*ConceptTag, error) {
	conceptTag := ConceptTag{}
	err := s.db.Where("tag=?", tag).First(&conceptTag).Error
	if err != nil {
		return nil, err
	}
	return &conceptTag, err
}

func (s *Store) ConceptTagsForConceptId(conceptId uint) ([]ConceptTag, error) {
	var conceptTags []ConceptTag
	err := s.db.Where("concept_id=?", conceptId).Order("concept_tags.order").Find(&conceptTags).Error
	return conceptTags, err
}

func (s *Store) ListConceptTags() ([]ConceptTag, error) {
	var conceptTags []ConceptTag
	err := s.db.Order("concept_tags.order").Find(&conceptTags).Error
	return conceptTags, err
}

func (s *Store) DeleteConceptTag(id uint) error {
	err := s.db.Unscoped().Where("id=?", id).Delete(ConceptTag{}).Error
	return err
}

func (s *Store) PurgeConceptTag(tag string) {
	s.db.Unscoped().Where("tag=?", tag).Delete(ConceptTag{})
}

func (s *Store) InsertItem(item *Item) (uint, error) {
	item.Location.Lat = float64(item.LatitudeI) / 10000000.0
	item.Location.Lng = float64(item.LongitudeI) / 10000000.0
	err := s.db.Create(item).Error
	return item.ID, err
}

func (s *Store) ListItemsByUser(userId uint, offset int, limit int) ([]Item, error) {
	var items []Item
	err := s.db.Where("user_id=?", userId).Offset(offset).Limit(limit).Find(&items).Error
	return items, err
}

func (s *Store) PurgeItem(itemId uint) error {
	return s.db.Unscoped().Where("id=?", itemId).Delete(Item{}).Error
}

func (s *Store) ListItemsForUserWithRegion(userId uint, regionId uint, offset int, limit int) ([]ItemWithOS, error) {
	permissions := UserPermissionsNone
	if userId > 0 {
		user, err := s.LoadPrivilegedUser(userId)
		if err != nil {
			return nil, err
		}
		permissions = user.Permissions
	}
	var err error
	var items []ItemWithOS
	err = s.db.Raw("SELECT items.*,ST_Transform(items.location, 27700) as osgb FROM regions LEFT JOIN items ON ST_Contains(regions.area, items.location) WHERE view_permissions<=? AND regions.id = ? ORDER BY silver_number OFFSET ? LIMIT ?", permissions, regionId, offset, limit).Scan(&items).Error
	if err != nil {
		return nil, err
	}
	return items, err
}

func (s *Store) ListItemsForUser(userId uint, offset int, limit int) ([]Item, error) {
	permissions := UserPermissionsNone
	if userId > 0 {
		user, err := s.LoadPrivilegedUser(userId)
		if err != nil {
			return nil, err
		}
		permissions = user.Permissions
	}
	var items []Item
	err := s.db.Where("view_permissions<=?", permissions).Offset(offset).Limit(limit).Order("silver_number").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, err
}

func (s *Store) LoadItem(id uint) (*Item, error) {
	item := Item{}
	err := s.db.Where("id=?", id).First(&item).Error
	return &item, err
}

func (s *Store) LoadItemBySilverNumber(id uint64) (*Item, error) {
	item := Item{}
	err := s.db.Where("silver_number=?", id).First(&item).Error
	return &item, err
}

func (s *Store) UpdateItem(item *Item) (uint, error) {
	item.Location.Lat = float64(item.LatitudeI) / 10000000.0
	item.Location.Lng = float64(item.LongitudeI) / 10000000.0
	err := s.db.Save(item).Error
	return item.ID, err
}

func (s *Store) InsertItemLog(itemLog *ItemLog) (uint, error) {
	err := s.db.Create(itemLog).Error
	return itemLog.ID, err
}

func (s *Store) ListItemLogsForItem(itemId uint) ([]ItemLog, error) {
	var itemLogs []ItemLog
	err := s.db.Where("item_id=?", itemId).Find(&itemLogs).Error
	return itemLogs, err
}

func (s *Store) ListItemLogs(userId uint, offset int, limit int) ([]ItemLog, error) {
	var itemLogs []ItemLog
	err := s.db.Offset(offset).Limit(limit).Order("created_at").Find(&itemLogs).Error
	if err != nil {
		return nil, err
	}
	return itemLogs, err
}

func (s *Store) LoadItemLogByUniqueId(unique_id string) (*ItemLog, error) {
	itemLog := ItemLog{}
	err := s.db.Where("unique_id=?", unique_id).First(&itemLog).Error
	return &itemLog, err
}
