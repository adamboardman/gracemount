module Types exposing (..)

import Array exposing (Array)
import Bootstrap.Modal as Modal
import Bootstrap.Navbar as Navbar
import Browser exposing (UrlRequest)
import Browser.Dom exposing (Element, Error)
import Browser.Navigation as Nav
import Char exposing (isDigit)
import Dict exposing (Dict)
import Dict.Extra exposing (fromListBy)
import FormatNumber exposing (format)
import FormatNumber.Locales exposing (Decimals(..), Locale, System(..))
import Http
import Json.Decode as Decode exposing (Decoder, field, float, int, list, map2, string)
import Json.Decode.Pipeline exposing (optional, required)
import Loading
import Set exposing (Set)
import String
import Time exposing (Month)
import Url exposing (Url)


type alias Model =
    { navKey : Maybe Nav.Key
    , page : Page
    , navState : Maybe Navbar.State
    , loading : Loading.LoadingState
    , problems : List Problem
    , loginForm : LoginForm
    , registerForm : RegisterForm
    , profileForm : ProfileForm
    , conceptForm : ConceptForm
    , conceptTagForm : ConceptTagForm
    , session : Session
    , apiActionResponse : ApiActionResponse
    , loggedInUser : User
    , concept : Concept
    , timeZone : Time.Zone
    , time : Time.Posix
    , date : Time.Posix
    , usersList : List User
    , conceptsList : List Concept
    , conceptTagsList : List ConceptTag
    , displayableTagsList : List DisplayableTag
    , conceptShowTagModel : Modal.Visibility
    , selectedItemId : Int
    , item : Item
    , itemLog : ItemLog
    , itemList : List Item
    , itemListFilter : String
    , itemLogList : List ItemLog
    , itemForm : ItemForm
    , itemLogForm : ItemLogForm
    , regionList : Maybe (List Region)
    , searchRegionId : Int
    , regionPlotWidth : Float
    }


type Page
    = Home
    | Login
    | Logout
    | Register (Maybe String) (Maybe String)
    | Profile
    | UsersList
    | UsersEdit String
    | Concepts String
    | ConceptsEdit String
    | ConceptsList
    | AddConcept
    | ItemList
    | AddItem
    | ItemEdit String
    | ItemLogList Int
      --| AddItemLog
      --| ItemLogEdit String
    | NotFound


type alias Session =
    { loginExpire : String
    , loginToken : String
    }


type alias ApiActionResponse =
    { status : Int
    , resourceId : Int
    , resourceIds : List Int
    }


type alias User =
    { id : Int
    , firstName : String
    , midNames : String
    , lastName : String
    , location : String
    , email : String
    , mobile : String
    , permissions : UserPermissionsType
    }


type alias Concept =
    { id : Int
    , name : String
    , summary : String
    , full : String
    , tags : List Tag
    }


type alias Item =
    { id : Int
    , userId : Int
    , viewPermissions : UserPermissionsType
    , silverNumber : Int
    , name : String
    , itemType : ItemTypeType
    , description : String
    , latitudeI : Int
    , longitudeI : Int
    , altitude : Float
    , status : String
    , ageGroup : String
    , height : Float
    , latinName : String
    , diameterAt : Float
    , renderHints : String
    , fruitType : String
    , croppingSeason : String
    , fruitStorage : String
    , pollinatingGroup : String
    , osgb : Point
    }



-- osgb - Point - lng == Northing, lat == Easting


type alias Point =
    { lng : Float
    , lat : Float
    }


type alias Region =
    { id : Int
    , name : String
    }


type alias ItemLog =
    { id : Int
    , date : Time.Posix
    , userId : Int
    , itemId : Int
    , name : String
    , description : String
    }


type alias Tag =
    { id : Int
    , order : Int
    , tag : String
    }


type alias ConceptTag =
    { id : Int
    , tag : String
    , conceptId : Int
    , order : Int
    }


type alias DisplayableTag =
    { id : Int
    , index : String
    , summary : String
    , tags : List String
    }


type alias LoginForm =
    { email : String
    , password : String
    }


type alias RegisterForm =
    { email : String
    , password : String
    , password_confirm : String
    , verification : String
    }


type alias ProfileForm =
    { id : Int
    , firstName : String
    , midNames : String
    , lastName : String
    , location : String
    , email : String
    , mobile : String
    , permissions : UserPermissionsType
    }


type alias ItemForm =
    { id : Int
    , userId : Int
    , viewPermissions : UserPermissionsType
    , silverNumber : String
    , name : String
    , itemType : ItemTypeType
    , description : String
    , latitude : String
    , longitude : String
    , altitude : String
    , status : String
    , ageGroup : String
    , height : String
    , latinName : String
    , diameterAt : String
    , renderHints : String
    , fruitType : String
    , croppingSeason : String
    , fruitStorage : String
    , pollinatingGroup : String
    }


type alias ItemLogForm =
    { id : Int
    , date : String
    , userId : Int
    , itemId : Int
    , name : String
    , description : String
    }


type alias ConceptForm =
    { name : String
    , tags : List ConceptTag
    , tagsToDelete : Set Int
    , summary : String
    , full : String
    }


type alias ConceptTagForm =
    { tag : String }


type alias CoordLatLon =
    { lat : Float
    , lon : Float
    }


type alias CoordBng =
    { easting : Float
    , northing : Float
    }


type ValidatedField
    = Email
    | Password
    | ConfirmPassword
    | FirstName
    | MidNames
    | LastName
    | Location
    | Mobile
    | Name
    | TagTag
    | Date
    | SilverNumber


type Problem
    = InvalidEntry ValidatedField String
    | ServerError String


type UserPermissionsType
    = UserPermissionsNone
    | UserPermissionsUser
    | UserPermissionsEditor
    | UserPermissionsAdmin
    | UserPermissionsMax


type ItemTypeType
    = ItemTypeAllOrUnknown
    | ItemTypeTree
    | ItemTypeTreeFruit
    | ItemTypeInfrastructure
    | ItemTypeMax


type Msg
    = ChangedUrl Url
    | ClickedLink UrlRequest
    | NavMsg Navbar.State
    | SubmittedLoginForm
    | SubmittedRegisterForm
    | SubmittedProfileForm
    | SubmittedConceptForm
    | SubmittedAddConceptTagForm
    | EnteredLoginEmail String
    | EnteredLoginPassword String
    | EnteredRegisterEmail String
    | EnteredRegisterPassword String
    | EnteredRegisterConfirmPassword String
    | EnteredUserFirstName String
    | EnteredUserMidNames String
    | EnteredUserLastName String
    | EnteredUserLocation String
    | EnteredUserMobile String
    | EnteredUserEmail String
    | SelectedUserPermissions UserPermissionsType
    | EnteredConceptName String
    | EnteredConceptTagCheckToDelete Int Bool
    | EnteredConceptSummary String
    | EnteredConceptFull String
    | EnteredAddConceptTag String
    | CompletedLogin (Result Http.Error Session)
    | GotRegisterJson (Result Http.Error ApiActionResponse)
    | LoadedUser (Result Http.Error User)
    | LoadedOtherUser (Result Http.Error User)
    | LoadedUsers (Result Http.Error (List User))
    | LoadedProfile (Result Http.Error ProfileForm)
    | LoadedConcept (Result Http.Error Concept)
    | LoadedConceptTags (Result Http.Error (List ConceptTag))
    | ConceptTagDeleted (Result Http.Error ApiActionResponse)
    | GotUpdateProfileJson (Result Http.Error ApiActionResponse)
    | AddedConcept (Result Http.Error ApiActionResponse)
    | AddedConceptTag Int String (Result Http.Error ApiActionResponse)
    | LoadedConcepts (Result Http.Error (List Concept))
    | LoadedConceptTagsList (Result Http.Error (List ConceptTag))
    | AdjustTimeZone Time.Zone
    | TimeTick Time.Posix
    | ButtonConceptAddTag
    | ButtonConceptDeleteSelectedTags
    | CloseConceptAddTagModal
    | EnteredFilter String
    | EnteredItemSilverNumber String
    | EnteredItemName String
    | SelectedItemType String
    | SelectedRegion String
    | EnteredItemDescription String
    | EnteredItemLatitude String
    | EnteredItemLongitude String
    | EnteredItemAltitude String
    | EnteredItemStatus String
    | EnteredItemAgeGroup String
    | EnteredItemHeight String
    | EnteredItemLatinName String
    | EnteredItemDiameterAt String
    | EnteredItemRenderHints String
    | EnteredItemFruitType String
    | EnteredItemCroppingSeason String
    | EnteredItemFruitStorage String
    | EnteredItemPollinatingGroup String
    | SelectedViewPermissions UserPermissionsType
      --| EnteredItemLogDate String
      --| EnteredItemLogName String
      --| EnteredItemLogDescription String
      --| SelectedItemLogUser String
      --| SelectedItemLogItem String
    | ViewItem Int
      --| ViewItemLog Int
    | SubmittedItemForm
      --| SubmittedItemLogForm
    | AddedItem (Result Http.Error ApiActionResponse)
      --| AddedItemLog (Result Http.Error ApiActionResponse)
    | LoadedItem (Result Http.Error Item)
    | LoadedItems (Result Http.Error (List Item))
    | LoadedItemLogs (Result Http.Error (List ItemLog))
    | LoadedRegions (Result Http.Error (List Region))
    | FilterByRegion
    | GotRegionPlotElement (Result Error Element)



-- FORMATTERS AND LOCALS


latLonLocale : Locale
latLonLocale =
    Locale (Exact 6) Western " " "." "−" "" "" "" "" ""


bngLocale : Locale
bngLocale =
    Locale (Exact 2) Western " " "." "−" "" "" "" "" ""


metersLocale : Locale
metersLocale =
    Locale (Exact 2) Western " " "." "−" "" "" "" "" ""


toIntMonth : Month -> Int
toIntMonth month =
    case month of
        Time.Jan ->
            1

        Time.Feb ->
            2

        Time.Mar ->
            3

        Time.Apr ->
            4

        Time.May ->
            5

        Time.Jun ->
            6

        Time.Jul ->
            7

        Time.Aug ->
            8

        Time.Sep ->
            9

        Time.Oct ->
            10

        Time.Nov ->
            11

        Time.Dec ->
            12


formatDate : Model -> Time.Posix -> String
formatDate model date =
    let
        year =
            String.fromInt (Time.toYear model.timeZone date)

        month =
            String.padLeft 2 '0' (String.fromInt (toIntMonth (Time.toMonth model.timeZone date)))

        day =
            String.padLeft 2 '0' (String.fromInt (Time.toDay model.timeZone date))
    in
    year ++ "-" ++ month ++ "-" ++ day


formatDateTime : Model -> Time.Posix -> String
formatDateTime model date =
    let
        year =
            String.fromInt (Time.toYear model.timeZone date)

        month =
            String.padLeft 2 '0' (String.fromInt (toIntMonth (Time.toMonth model.timeZone date)))

        day =
            String.padLeft 2 '0' (String.fromInt (Time.toDay model.timeZone date))

        hour =
            String.padLeft 2 '0' (String.fromInt (Time.toHour model.timeZone date))

        minute =
            String.padLeft 2 '0' (String.fromInt (Time.toMinute model.timeZone date))
    in
    year ++ "-" ++ month ++ "-" ++ day ++ " " ++ hour ++ ":" ++ minute


formatSilverNumber : Int -> String
formatSilverNumber num =
    String.padLeft 4 '0' (String.fromInt num)


formatLatLon : CoordLatLon -> String
formatLatLon coord =
    "Lat: " ++ format latLonLocale coord.lat ++ ", Lon: " ++ format latLonLocale coord.lon ++ " "


formatBng : CoordBng -> String
formatBng coord =
    "Easting: " ++ format bngLocale coord.easting ++ ", Northing: " ++ format bngLocale coord.northing ++ " "


formatDiameterAt : Float -> String
formatDiameterAt diameter =
    let
        m =
            diameter * pi

        cm =
            m * 100.0
    in
    if cm > 100 then
        "Circumference At 1m: " ++ format metersLocale m ++ "m"

    else
        "Circumference At 1m: " ++ String.fromInt (round cm) ++ "cm"


formatHeight : Float -> String
formatHeight height =
    "Height: " ++ format metersLocale height ++ "m"


formatAltitude : Float -> String
formatAltitude height =
    "Altitude: " ++ format metersLocale height ++ "m"


secondsFromTime : String -> Int
secondsFromTime time =
    let
        timeParts =
            Array.fromList (String.split ":" (timeFromTime time))

        hours =
            Maybe.withDefault 0 (String.toInt (Maybe.withDefault "0" (Array.get 0 timeParts)))

        minutes =
            Maybe.withDefault 0 (String.toInt (Maybe.withDefault "0" (Array.get 1 timeParts)))

        seconds =
            Maybe.withDefault 0 (String.toInt (Maybe.withDefault "0" (Array.get 2 timeParts)))
    in
    (hours * (60 * 60)) + (minutes * 60) + seconds


secondsFromTimeHMS : String -> String -> String -> Int
secondsFromTimeHMS timeH timeM timeS =
    let
        hours =
            Maybe.withDefault 0 (String.toInt timeH)

        minutes =
            Maybe.withDefault 0 (String.toInt timeM)

        seconds =
            Maybe.withDefault 0 (String.toInt timeS)
    in
    (hours * (60 * 60)) + (minutes * 60) + seconds


timeFromTime : String -> String
timeFromTime time =
    let
        timeParts =
            Array.fromList (String.split ":" time)

        hours =
            String.padLeft 2 '0' (Maybe.withDefault "" (Array.get 0 timeParts))

        minutes =
            String.padLeft 2 '0' (String.slice 0 2 (Maybe.withDefault "" (Array.get 1 timeParts)))

        seconds =
            String.padLeft 2 '0' (String.slice 0 2 (Maybe.withDefault "" (Array.get 2 timeParts)))
    in
    hours ++ ":" ++ minutes ++ ":" ++ seconds


isDigitOrPlace : Char -> Bool
isDigitOrPlace char =
    if isDigit char || char == '.' then
        True

    else
        False


isNot : Int -> Int -> Bool
isNot a b =
    if a == b then
        False

    else
        True


dateFromItemLog : Model -> ItemLog -> String
dateFromItemLog model log =
    if Time.posixToMillis log.date > 0 then
        formatDateTime model log.date

    else
        ""



-- INDEXERS


indexUser : User -> ( String, User )
indexUser user =
    ( String.fromInt user.id, user )


idFromConcept : Concept -> Int
idFromConcept concept =
    concept.id


idFromDisplayable : DisplayableTag -> Int
idFromDisplayable dTag =
    dTag.id


conceptIdFromConceptTag : ConceptTag -> Int
conceptIdFromConceptTag conceptTag =
    conceptTag.conceptId


tagFromConceptTagIfMatching : Int -> ConceptTag -> Maybe String
tagFromConceptTagIfMatching conceptId conceptTag =
    if conceptTag.conceptId == conceptId then
        Just conceptTag.tag

    else
        Nothing


displayableTagFrom : List ConceptTag -> Dict Int Concept -> Int -> DisplayableTag
displayableTagFrom conceptTags concepts conceptId =
    let
        tags =
            List.filterMap (tagFromConceptTagIfMatching conceptId) conceptTags

        index =
            case List.head tags of
                Just tag ->
                    tag

                Nothing ->
                    ""

        maybeConcept =
            Dict.get conceptId concepts

        summary =
            case maybeConcept of
                Just concept ->
                    concept.summary

                Nothing ->
                    ""
    in
    { id = conceptId
    , index = index
    , summary = summary
    , tags = tags
    }


displayableTagsListFrom : List ConceptTag -> List Concept -> List DisplayableTag
displayableTagsListFrom conceptTags concepts =
    let
        conceptIdList =
            Set.toList (Set.fromList (List.map conceptIdFromConceptTag conceptTags))

        groupedConcepts =
            fromListBy idFromConcept concepts

        dTags =
            List.map (displayableTagFrom conceptTags groupedConcepts) conceptIdList
    in
    dTags



-- EMPTIES


emptyUser : User
emptyUser =
    { id = 0
    , firstName = ""
    , midNames = ""
    , lastName = ""
    , location = ""
    , email = ""
    , mobile = ""
    , permissions = UserPermissionsNone
    }


emptySession : Session
emptySession =
    { loginExpire = "", loginToken = "" }


emptyConcept : Concept
emptyConcept =
    { id = 0
    , name = ""
    , summary = ""
    , full = ""
    , tags = []
    }


emptyConceptForm : ConceptForm
emptyConceptForm =
    { name = ""
    , tags = []
    , tagsToDelete = Set.empty
    , summary = ""
    , full = ""
    }


emptyProfileForm : ProfileForm
emptyProfileForm =
    { id = 0
    , firstName = ""
    , midNames = ""
    , lastName = ""
    , location = ""
    , email = ""
    , mobile = ""
    , permissions = UserPermissionsNone
    }


emptyItemForm : ItemForm
emptyItemForm =
    { id = 0
    , userId = 0
    , viewPermissions = UserPermissionsNone
    , silverNumber = ""
    , name = ""
    , itemType = ItemTypeAllOrUnknown
    , description = ""
    , latitude = ""
    , longitude = ""
    , altitude = ""
    , status = ""
    , ageGroup = ""
    , height = ""
    , latinName = ""
    , diameterAt = ""
    , renderHints = ""
    , fruitType = ""
    , croppingSeason = ""
    , fruitStorage = ""
    , pollinatingGroup = ""
    }


emptyPoint : Point
emptyPoint =
    { lng = 0
    , lat = 0
    }


emptyItem : Item
emptyItem =
    { id = 0
    , userId = 0
    , viewPermissions = UserPermissionsNone
    , silverNumber = 0
    , name = ""
    , itemType = ItemTypeAllOrUnknown
    , description = ""
    , latitudeI = 0
    , longitudeI = 0
    , altitude = 0
    , status = ""
    , ageGroup = ""
    , height = 0
    , latinName = ""
    , diameterAt = 0
    , renderHints = ""
    , fruitType = ""
    , croppingSeason = ""
    , fruitStorage = ""
    , pollinatingGroup = ""
    , osgb = emptyPoint
    }


emptyItemLog : ItemLog
emptyItemLog =
    { id = 0
    , date = Time.millisToPosix 0
    , userId = 0
    , itemId = 0
    , name = ""
    , description = ""
    }


emptyItemLogForm : ItemLogForm
emptyItemLogForm =
    { id = 0
    , date = ""
    , userId = 0
    , itemId = 0
    , name = ""
    , description = ""
    }



-- DECODERS


resourceIdsDecoder : Decoder (List Int)
resourceIdsDecoder =
    list int


apiActionDecoder : Decoder ApiActionResponse
apiActionDecoder =
    Decode.succeed ApiActionResponse
        |> required "status" int
        |> optional "resourceId" int 0
        |> optional "resourceIds" resourceIdsDecoder []


permissionsFromInt : Int -> UserPermissionsType
permissionsFromInt int =
    case int of
        0 ->
            UserPermissionsNone

        1 ->
            UserPermissionsUser

        2 ->
            UserPermissionsEditor

        3 ->
            UserPermissionsAdmin

        _ ->
            UserPermissionsMax


userPermissionsDecoder : Decoder UserPermissionsType
userPermissionsDecoder =
    Decode.int
        |> Decode.andThen
            (\modeInt ->
                Decode.succeed (permissionsFromInt modeInt)
            )


userPermissionsToInt : UserPermissionsType -> Int
userPermissionsToInt permissions =
    case permissions of
        UserPermissionsNone ->
            0

        UserPermissionsUser ->
            1

        UserPermissionsEditor ->
            2

        UserPermissionsAdmin ->
            3

        UserPermissionsMax ->
            4


itemTypeFromInt : Int -> ItemTypeType
itemTypeFromInt int =
    case int of
        0 ->
            ItemTypeAllOrUnknown

        1 ->
            ItemTypeTree

        2 ->
            ItemTypeTreeFruit

        3 ->
            ItemTypeInfrastructure

        _ ->
            ItemTypeMax


itemTypeDecoder : Decoder ItemTypeType
itemTypeDecoder =
    Decode.int
        |> Decode.andThen
            (\modeInt ->
                Decode.succeed (itemTypeFromInt modeInt)
            )


itemTypeToInt : ItemTypeType -> Int
itemTypeToInt itemType =
    case itemType of
        ItemTypeAllOrUnknown ->
            0

        ItemTypeTree ->
            1

        ItemTypeTreeFruit ->
            2

        ItemTypeInfrastructure ->
            3

        ItemTypeMax ->
            4


itemTypeToString : ItemTypeType -> String
itemTypeToString itemType =
    case itemType of
        ItemTypeAllOrUnknown ->
            "All / Unknown"

        ItemTypeTree ->
            "Tree"

        ItemTypeTreeFruit ->
            "Fruit Tree"

        ItemTypeInfrastructure ->
            "Infrastructure"

        ItemTypeMax ->
            "[Should not be displayed!]"


osgbDecoder : Decoder Point
osgbDecoder =
    map2 Point
        (field "lng" float)
        (field "lat" float)


userPermissionsToString : UserPermissionsType -> String
userPermissionsToString permissions =
    case permissions of
        UserPermissionsNone ->
            "None"

        UserPermissionsUser ->
            "User"

        UserPermissionsEditor ->
            "Editor"

        UserPermissionsAdmin ->
            "Admin"

        UserPermissionsMax ->
            "[Should not be displayed!]"


userDecoder : Decoder User
userDecoder =
    Decode.succeed User
        |> required "ID" int
        |> required "FirstName" string
        |> required "MidNames" string
        |> required "LastName" string
        |> optional "Location" string ""
        |> optional "Email" string ""
        |> optional "Mobile" string ""
        |> optional "Permissions" userPermissionsDecoder UserPermissionsNone


profileDecoder : Decoder ProfileForm
profileDecoder =
    Decode.succeed ProfileForm
        |> required "ID" int
        |> required "FirstName" string
        |> required "MidNames" string
        |> required "LastName" string
        |> required "Location" string
        |> required "Email" string
        |> required "Mobile" string
        |> required "Permissions" userPermissionsDecoder


conceptDecoder : Decoder Concept
conceptDecoder =
    Decode.succeed Concept
        |> required "ID" int
        |> required "Name" string
        |> required "Summary" string
        |> required "Full" string
        |> optional "Tags" (list tagDecoder) []


conceptTagsListDecoder : Decoder (List ConceptTag)
conceptTagsListDecoder =
    list conceptTagDecoder


tagDecoder : Decoder Tag
tagDecoder =
    Decode.succeed Tag
        |> required "ID" int
        |> required "Order" int
        |> required "Tag" string


conceptTagDecoder : Decoder ConceptTag
conceptTagDecoder =
    Decode.succeed ConceptTag
        |> required "ID" int
        |> required "Tag" string
        |> required "ConceptId" int
        |> required "Order" int


posixTime : Decode.Decoder Time.Posix
posixTime =
    Decode.int
        |> Decode.andThen
            (\ms -> Decode.succeed <| Time.millisToPosix ms)


itemDecoder : Decoder Item
itemDecoder =
    Decode.succeed Item
        |> required "ID" int
        |> required "UserId" int
        |> required "ViewPermissions" userPermissionsDecoder
        |> required "SilverNumber" int
        |> required "Name" string
        |> required "ItemType" itemTypeDecoder
        |> required "Description" string
        |> required "LatitudeI" int
        |> required "LongitudeI" int
        |> required "Altitude" float
        |> required "Status" string
        |> required "AgeGroup" string
        |> required "Height" float
        |> required "LatinName" string
        |> required "DiameterAt" float
        |> required "RenderHints" string
        |> required "FruitType" string
        |> required "CroppingSeason" string
        |> required "FruitStorage" string
        |> required "PollinatingGroup" string
        |> optional "OSGB" osgbDecoder emptyPoint


regionDecoder : Decoder Region
regionDecoder =
    Decode.succeed Region
        |> required "ID" int
        |> required "Name" string


itemLogDecoder : Decoder ItemLog
itemLogDecoder =
    Decode.succeed ItemLog
        |> required "ID" int
        |> required "Date" posixTime
        |> required "UserId" int
        |> required "ItemId" int
        |> required "Name" string
        |> required "Description" string



-- AUTH HEADER


authHeader : String -> Http.Header
authHeader token =
    Http.header "authorization" ("Bearer " ++ token)
