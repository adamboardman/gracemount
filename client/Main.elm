module Main exposing (main)

import Bootstrap.Grid as Grid
import Bootstrap.Modal as Modal
import Bootstrap.Navbar as Navbar
import Browser exposing (Document, UrlRequest)
import Browser.Dom
import Browser.Navigation as Nav
import Concept exposing (pageConcept)
import ConceptsEdit exposing (conceptAdd, conceptDeleteSelectedTags, conceptTag, conceptTagUpdateForm, conceptTagValidate, conceptUpdate, conceptUpdateForm, conceptValidate, loadConceptById, loadConceptTagsById, pageAddConcept, pageConceptsEdit, tagIsNotIn)
import ConceptsList exposing (loadConceptTagsList, loadConcepts, pageConceptsList)
import Date
import Html exposing (Html, div, h1, text)
import Html.Attributes exposing (href)
import Http exposing (Error(..), emptyBody)
import Iso8601
import ItemEdit exposing (itemAdd, itemUpdate, itemUpdateForm, pageAddItem, pageItemEdit)
import ItemList exposing (loadItemById, loadItems, loadItemsForRegion, loadRegions, pageItemList)
import ItemLogList exposing (loadItemLogs, pageItemLogList)
import Loading
import Login exposing (loggedIn, login, loginUpdateForm, loginValidate, pageLogin, userIsAdmin, userIsEditor)
import Ports exposing (storeExpire, storeToken)
import Profile exposing (pageProfile, pageUsersEdit, profile, profileUpdateForm, profileValidate)
import Register exposing (pageRegister, register, registerUpdateForm, registerValidate)
import Set
import Task
import Time exposing (utc)
import Types exposing (LoginForm, Model, Msg(..), Page(..), Problem(..), Session, User, authHeader, conceptDecoder, displayableTagsListFrom, emptyConcept, emptyConceptForm, emptyItem, emptyItemForm, emptyItemLog, emptyItemLogForm, emptyProfileForm, emptySession, emptyUser, indexUser, isNot, itemTypeFromInt, profileDecoder, userDecoder)
import Url exposing (Url)
import Url.Parser as UrlParser exposing ((</>), (<?>), Parser, int, s, string, top)
import Url.Parser.Query as Query
import UsersList exposing (loadUsers, pageUsersList)



-- TYPES


type alias Flags =
    { token : Maybe String
    , expire : Maybe String
    }



-- MAIN


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , onUrlChange = ChangedUrl
        , onUrlRequest = ClickedLink
        , subscriptions = subscriptions
        , update = update
        , view = view
        }


init : Flags -> Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        ( navState, navCmd ) =
            Navbar.initialState NavMsg

        ( model, urlCmd ) =
            urlUpdate url
                { navKey = Just key
                , navState = Just navState
                , page = Home
                , loading = Loading.Off
                , problems = []
                , usersList = []
                , loginForm = { email = "", password = "" }
                , registerForm = { email = "", password = "", password_confirm = "", verification = "" }
                , session = { loginExpire = Maybe.withDefault "" flags.expire, loginToken = Maybe.withDefault "" flags.token }
                , apiActionResponse = { status = 0, resourceId = 0, resourceIds = [] }
                , loggedInUser = emptyUser
                , profileForm = emptyProfileForm
                , conceptForm = emptyConceptForm
                , conceptTagForm = { tag = "" }
                , concept = emptyConcept
                , timeZone = Time.utc
                , time = Time.millisToPosix 0
                , date = Time.millisToPosix 0
                , conceptsList = []
                , conceptTagsList = []
                , displayableTagsList = []
                , conceptShowTagModel = Modal.hidden
                , selectedItemId = 0
                , item = emptyItem
                , itemForm = emptyItemForm
                , itemList = []
                , itemListFilter = ""
                , itemLog = emptyItemLog
                , itemLogForm = emptyItemLogForm
                , itemLogList = []
                , regionList = Nothing
                , searchRegionId = 0
                , regionPlotWidth = 1000
                }
    in
    ( model
    , Cmd.batch
        [ urlCmd
        , navCmd
        , case flags.token of
            Just token ->
                loadUser token 0

            Nothing ->
                Cmd.none
        , Task.perform AdjustTimeZone Time.here
        , Task.perform TimeTick Time.now
        ]
    )


view : Model -> Document Msg
view model =
    { title = "Gracemount - Tree and Plant Catalogue"
    , body =
        [ div []
            [ menu model
            , mainContent model
            ]
        ]
    }


menu : Model -> Html Msg
menu model =
    case model.navState of
        Just navState ->
            Navbar.config NavMsg
                |> Navbar.withAnimation
                --|> Navbar.container
                |> Navbar.brand [ href (urlForPage Home) ] [ text "Gracemount" ]
                |> Navbar.items
                    [ if userIsEditor model then
                        Navbar.dropdown
                            { id = "concepts_dropdown"
                            , toggle = Navbar.dropdownToggle [] [ text "Concepts" ]
                            , items =
                                [ Navbar.dropdownItem
                                    [ href (urlForPage ConceptsList) ]
                                    [ text "Concepts" ]
                                , Navbar.dropdownItem
                                    [ href (urlForPage AddConcept) ]
                                    [ text "Add Concept" ]
                                ]
                            }

                      else
                        Navbar.itemLink [] [ text "" ]
                    , if userIsEditor model then
                        Navbar.dropdown
                            { id = "items_dropdown"
                            , toggle = Navbar.dropdownToggle [] [ text "Items" ]
                            , items =
                                [ Navbar.dropdownItem
                                    [ href (urlForPage ItemList) ]
                                    [ text "Items" ]
                                , Navbar.dropdownItem
                                    [ href (urlForPage AddItem) ]
                                    [ text "Add Item" ]

                                --, Navbar.dropdownItem
                                --    [ href (urlForPage ItemLogList) ]
                                --    [ text "Item Logs" ]
                                --, Navbar.dropdownItem
                                --    [ href (urlForPage AddItemLog) ]
                                --    [ text "Add Item Log" ]
                                ]
                            }

                      else
                        Navbar.itemLink [] [ text "" ]
                    , if loggedIn model then
                        if userIsAdmin model then
                            Navbar.dropdown
                                { id = "users_dropdown"
                                , toggle = Navbar.dropdownToggle [] [ text "Users" ]
                                , items =
                                    [ Navbar.dropdownItem
                                        [ href (urlForPage UsersList) ]
                                        [ text "Users" ]
                                    , Navbar.dropdownItem
                                        [ href (urlForPage Profile) ]
                                        [ text "Profile" ]
                                    ]
                                }

                        else
                            Navbar.itemLink [ href (urlForPage Profile) ] [ text "Profile" ]

                      else
                        Navbar.itemLink [] [ text "" ]
                    , if loggedIn model then
                        Navbar.itemLink [ href (urlForPage Logout) ] [ text "Logout" ]

                      else
                        Navbar.itemLink [ href (urlForPage Login) ] [ text "Login" ]
                    ]
                |> Navbar.view navState

        Nothing ->
            div [] []


mainContent : Model -> Html Msg
mainContent model =
    Grid.container [] <|
        case model.page of
            Home ->
                pageConcept model

            Login ->
                pageLogin model

            Logout ->
                pageLogout model

            Register email _ ->
                pageRegister model email

            Profile ->
                pageProfile model

            UsersList ->
                pageUsersList model

            Concepts _ ->
                pageConcept model

            ConceptsList ->
                pageConceptsList model

            ConceptsEdit _ ->
                pageConceptsEdit model

            AddConcept ->
                pageAddConcept model

            ItemList ->
                pageItemList model

            AddItem ->
                pageAddItem model

            --AddItemLog ->
            --    pageAddItemLog model
            ItemLogList id ->
                pageItemLogList model id

            --ItemLogEdit _ ->
            --    pageItemLogEdit model
            ItemEdit _ ->
                pageItemEdit model

            UsersEdit _ ->
                pageUsersEdit model

            NotFound ->
                pageNotFound


pageLogout : Model -> List (Html Msg)
pageLogout model =
    [ text "Logged Out"
    ]


pageNotFound : List (Html Msg)
pageNotFound =
    [ h1 [] [ text "Not found" ]
    , text "Sorry couldn't find that page"
    ]



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ClickedLink urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( { model
                        | problems = []
                        , loginForm = { email = "", password = "" }
                        , registerForm = { email = "", password = "", password_confirm = "", verification = "" }
                        , apiActionResponse = { status = 0, resourceId = 0, resourceIds = [] }
                        , session =
                            case url.path of
                                "/logout" ->
                                    emptySession

                                _ ->
                                    model.session
                      }
                    , case model.navKey of
                        Just navKey ->
                            case url.path of
                                "/logout" ->
                                    Cmd.batch
                                        [ Nav.pushUrl navKey (urlForPage Home)
                                        , storeToken Nothing
                                        , storeExpire Nothing
                                        ]

                                _ ->
                                    Nav.pushUrl navKey (Url.toString url)

                        Nothing ->
                            Cmd.none
                    )

                Browser.External href ->
                    ( model, Nav.load href )

        ChangedUrl url ->
            urlUpdate url model

        NavMsg state ->
            ( { model | navState = Just state }, Cmd.none )

        CloseConceptAddTagModal ->
            ( { model | conceptShowTagModel = Modal.hidden }, Cmd.none )

        SubmittedLoginForm ->
            case loginValidate model.loginForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , login validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        SubmittedRegisterForm ->
            case registerValidate model.registerForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , register validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        SubmittedProfileForm ->
            case profileValidate model.profileForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , profile model.session.loginToken validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        SubmittedConceptForm ->
            case conceptValidate model.conceptForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , if model.concept.id > 0 then
                        conceptUpdate model validForm

                      else
                        conceptAdd model validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        SubmittedAddConceptTagForm ->
            case conceptTagValidate model.conceptTagForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , conceptTag model validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        EnteredLoginEmail email ->
            loginUpdateForm (\form -> { form | email = email }) model

        EnteredRegisterEmail email ->
            registerUpdateForm (\form -> { form | email = email }) model

        EnteredLoginPassword password ->
            loginUpdateForm (\form -> { form | password = password }) model

        EnteredRegisterPassword password ->
            registerUpdateForm (\form -> { form | password = password }) model

        EnteredRegisterConfirmPassword passwordConfirm ->
            registerUpdateForm (\form -> { form | password_confirm = passwordConfirm }) model

        EnteredUserFirstName firstName ->
            profileUpdateForm (\form -> { form | firstName = firstName }) model

        EnteredUserMidNames midNames ->
            profileUpdateForm (\form -> { form | midNames = midNames }) model

        EnteredUserLastName lastName ->
            profileUpdateForm (\form -> { form | lastName = lastName }) model

        EnteredUserLocation location ->
            profileUpdateForm (\form -> { form | location = location }) model

        EnteredUserMobile mobile ->
            profileUpdateForm (\form -> { form | mobile = mobile }) model

        EnteredUserEmail email ->
            profileUpdateForm (\form -> { form | email = email }) model

        SelectedUserPermissions permissions ->
            profileUpdateForm (\form -> { form | permissions = permissions }) model

        EnteredConceptName name ->
            conceptUpdateForm (\form -> { form | name = name }) model

        EnteredConceptTagCheckToDelete tagId tagState ->
            let
                tags =
                    if tagState then
                        Set.insert tagId model.conceptForm.tagsToDelete

                    else
                        Set.filter (isNot tagId) model.conceptForm.tagsToDelete
            in
            conceptUpdateForm (\form -> { form | tagsToDelete = tags }) model

        EnteredConceptSummary summary ->
            conceptUpdateForm (\form -> { form | summary = summary }) model

        EnteredConceptFull full ->
            conceptUpdateForm (\form -> { form | full = full }) model

        EnteredAddConceptTag tag ->
            conceptTagUpdateForm (\form -> { form | tag = tag }) model

        ButtonConceptAddTag ->
            ( { model | conceptShowTagModel = Modal.shown }, Cmd.none )

        ButtonConceptDeleteSelectedTags ->
            ( { model | loading = Loading.On }, conceptDeleteSelectedTags model )

        CompletedLogin (Err error) ->
            let
                serverErrors =
                    decodeErrors error
                        |> List.map ServerError
            in
            ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off }
            , Cmd.batch
                [ storeToken Nothing
                , storeExpire Nothing
                ]
            )

        CompletedLogin (Ok res) ->
            ( { model | session = res, loading = Loading.Off }
            , Cmd.batch
                [ loadUser res.loginToken 0
                , storeToken (Just res.loginToken)
                , storeExpire (Just res.loginExpire)
                , case model.navKey of
                    Just navKey ->
                        Nav.pushUrl navKey (urlForPage Home)

                    Nothing ->
                        Cmd.none
                ]
            )

        LoadedUser (Err error) ->
            ( { model | loggedInUser = emptyUser, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedUser (Ok res) ->
            ( { model | loggedInUser = res, loading = Loading.Off }
            , Cmd.none
            )

        LoadedOtherUser (Err error) ->
            ( { model | profileForm = emptyProfileForm, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedOtherUser (Ok res) ->
            let
                profileForm =
                    { id = res.id
                    , firstName = res.firstName
                    , midNames = res.midNames
                    , lastName = res.lastName
                    , location = res.location
                    , email = res.email
                    , mobile = res.mobile
                    , permissions = res.permissions
                    }
            in
            ( { model | profileForm = profileForm, loading = Loading.Off }
            , Cmd.none
            )

        LoadedUsers (Err error) ->
            ( { model | loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedUsers (Ok res) ->
            ( { model | usersList = res, loading = Loading.Off }
            , Cmd.none
            )

        LoadedProfile (Err error) ->
            ( { model | profileForm = emptyProfileForm, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedProfile (Ok res) ->
            let
                profileForm =
                    { id = res.id
                    , firstName = res.firstName
                    , midNames = res.midNames
                    , lastName = res.lastName
                    , location = res.location
                    , email = res.email
                    , mobile = res.mobile
                    , permissions = res.permissions
                    }
            in
            ( { model | profileForm = profileForm, loading = Loading.Off }
            , Cmd.none
            )

        LoadedConcept (Err error) ->
            ( { model
                | concept = emptyConcept
                , conceptForm = emptyConceptForm
                , loading = Loading.Off
                , session = sessionGivenAuthError error model
              }
            , Cmd.none
            )

        LoadedConcept (Ok res) ->
            let
                conceptForm =
                    { name = res.name
                    , tags = model.conceptForm.tags
                    , tagsToDelete = Set.empty
                    , summary = res.summary
                    , full = res.full
                    }
            in
            ( { model | concept = res, conceptForm = conceptForm, loading = Loading.Off }
            , Cmd.none
            )

        LoadedConceptTags (Err error) ->
            let
                serverErrors =
                    decodeErrors error
                        |> List.map ServerError
            in
            ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedConceptTags (Ok res) ->
            let
                conceptForm =
                    { name = model.conceptForm.name
                    , tags = res
                    , tagsToDelete = model.conceptForm.tagsToDelete
                    , summary = model.conceptForm.summary
                    , full = model.conceptForm.full
                    }
            in
            ( { model | conceptForm = conceptForm, problems = [], loading = Loading.Off }
            , Cmd.none
            )

        LoadedConcepts (Err error) ->
            ( { model | loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedConcepts (Ok res) ->
            let
                dTags =
                    displayableTagsListFrom model.conceptTagsList res
            in
            ( { model | conceptsList = res, displayableTagsList = dTags, loading = Loading.Off }
            , Cmd.none
            )

        LoadedConceptTagsList (Err error) ->
            let
                serverErrors =
                    decodeErrors error
                        |> List.map ServerError
            in
            ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedConceptTagsList (Ok res) ->
            let
                dTags =
                    displayableTagsListFrom res model.conceptsList
            in
            ( { model | conceptTagsList = res, displayableTagsList = dTags, loading = Loading.Off }
            , Cmd.none
            )

        ConceptTagDeleted (Err error) ->
            let
                serverErrors =
                    decodeErrors error
                        |> List.map ServerError
            in
            ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        ConceptTagDeleted (Ok res) ->
            let
                conceptFormTags =
                    List.filter (tagIsNotIn (Set.fromList res.resourceIds)) model.conceptForm.tags

                conceptForm =
                    { name = model.conceptForm.name
                    , summary = model.conceptForm.summary
                    , full = model.conceptForm.full
                    , tagsToDelete = model.conceptForm.tagsToDelete
                    , tags = conceptFormTags
                    }
            in
            ( { model | conceptForm = conceptForm, loading = Loading.Off }
            , Cmd.none
            )

        GotRegisterJson result ->
            case result of
                Ok res ->
                    ( { model | apiActionResponse = res, loading = Loading.Off }, Cmd.none )

                Err error ->
                    let
                        serverErrors =
                            decodeErrors error
                                |> List.map ServerError
                    in
                    ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
                    , Cmd.none
                    )

        GotUpdateProfileJson result ->
            case result of
                Ok res ->
                    ( { model | apiActionResponse = res, loading = Loading.Off }, Cmd.none )

                Err error ->
                    let
                        serverErrors =
                            decodeErrors error
                                |> List.map ServerError
                    in
                    ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
                    , Cmd.none
                    )

        AddedConcept result ->
            case result of
                Ok res ->
                    ( { model | apiActionResponse = res, loading = Loading.Off }
                    , loadConceptById res.resourceId
                    )

                Err error ->
                    let
                        serverErrors =
                            decodeErrors error
                                |> List.map ServerError

                        newSession =
                            sessionGivenAuthError error model
                    in
                    ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = newSession }
                    , Cmd.none
                    )

        AddedConceptTag conceptId tag result ->
            case result of
                Ok res ->
                    let
                        conceptTag =
                            { id = res.resourceId, tag = tag, conceptId = conceptId, order = 0 }

                        conceptFormTags =
                            model.conceptForm.tags ++ [ conceptTag ]

                        conceptForm =
                            { name = model.conceptForm.name
                            , summary = model.conceptForm.summary
                            , full = model.conceptForm.full
                            , tagsToDelete = model.conceptForm.tagsToDelete
                            , tags = conceptFormTags
                            }

                        tags =
                            model.concept.tags
                                ++ [ { id = res.resourceId
                                     , order = 0
                                     , tag = tag
                                     }
                                   ]

                        concept =
                            { id = model.concept.id
                            , name = model.concept.name
                            , summary = model.concept.summary
                            , full = model.concept.full
                            , tags = tags
                            }
                    in
                    ( { model | concept = concept, conceptForm = conceptForm, apiActionResponse = res, loading = Loading.Off }
                    , Cmd.none
                    )

                Err error ->
                    let
                        serverErrors =
                            decodeErrors error
                                |> List.map ServerError

                        newSession =
                            sessionGivenAuthError error model
                    in
                    ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = newSession }
                    , Cmd.none
                    )

        AdjustTimeZone zone ->
            ( { model | timeZone = zone }, Cmd.none )

        TimeTick posix ->
            let
                dateIso =
                    String.fromInt (Date.year (Date.fromPosix utc posix))

                zeroDateTime =
                    Iso8601.fromTime (Time.millisToPosix 0)

                zeroTime =
                    String.slice (String.length dateIso) (String.length zeroDateTime) zeroDateTime

                dateZeroTime =
                    dateIso ++ zeroTime

                dateResult =
                    Iso8601.toTime dateZeroTime

                dateTime =
                    case dateResult of
                        Ok time ->
                            time

                        _ ->
                            model.time
            in
            ( { model | time = posix, date = dateTime }, Cmd.none )

        LoadedItems (Err error) ->
            ( { model | loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedItems (Ok res) ->
            ( { model | itemList = res, loading = Loading.Off }
            , Cmd.none
            )

        LoadedRegions (Err error) ->
            ( { model | regionList = Nothing, loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedRegions (Ok res) ->
            ( { model | regionList = Just res, loading = Loading.Off }
            , Cmd.none
            )

        LoadedItemLogs (Err error) ->
            ( { model | loading = Loading.Off, session = sessionGivenAuthError error model }
            , Cmd.none
            )

        LoadedItemLogs (Ok res) ->
            ( { model | itemLogList = res, loading = Loading.Off }
            , Cmd.none
            )

        EnteredFilter string ->
            ( { model | itemListFilter = string }
            , Cmd.none
            )

        SelectedRegion string ->
            let
                newModel =
                    { model | searchRegionId = Maybe.withDefault 0 (String.toInt string), page = ItemList }
            in
            ( newModel
            , Cmd.batch [ loadItemsForRegion newModel, loadRegions newModel ]
            )

        EnteredItemSilverNumber string ->
            itemUpdateForm (\form -> { form | silverNumber = string }) model

        EnteredItemName string ->
            itemUpdateForm (\form -> { form | name = string }) model

        SelectedItemType string ->
            itemUpdateForm (\form -> { form | itemType = itemTypeFromInt (Maybe.withDefault 0 (String.toInt string)) }) model

        EnteredItemDescription string ->
            itemUpdateForm (\form -> { form | description = string }) model

        EnteredItemLatitude string ->
            itemUpdateForm (\form -> { form | latitude = string }) model

        EnteredItemLongitude string ->
            itemUpdateForm (\form -> { form | longitude = string }) model

        EnteredItemAltitude string ->
            itemUpdateForm (\form -> { form | altitude = string }) model

        EnteredItemStatus string ->
            itemUpdateForm (\form -> { form | status = string }) model

        EnteredItemAgeGroup string ->
            itemUpdateForm (\form -> { form | ageGroup = string }) model

        EnteredItemHeight string ->
            itemUpdateForm (\form -> { form | height = string }) model

        EnteredItemLatinName string ->
            itemUpdateForm (\form -> { form | latinName = string }) model

        EnteredItemDiameterAt string ->
            itemUpdateForm (\form -> { form | diameterAt = string }) model

        EnteredItemRenderHints string ->
            itemUpdateForm (\form -> { form | renderHints = string }) model

        EnteredItemFruitType string ->
            itemUpdateForm (\form -> { form | fruitType = string }) model

        EnteredItemCroppingSeason string ->
            itemUpdateForm (\form -> { form | croppingSeason = string }) model

        EnteredItemFruitStorage string ->
            itemUpdateForm (\form -> { form | fruitStorage = string }) model

        EnteredItemPollinatingGroup string ->
            itemUpdateForm (\form -> { form | pollinatingGroup = string }) model

        SelectedViewPermissions int ->
            itemUpdateForm (\form -> { form | viewPermissions = int }) model

        --EnteredItemLogDate string ->
        --    itemLogUpdateForm (\form -> { form | date = string }) model
        --
        --
        --EnteredItemLogName string ->
        --    itemLogUpdateForm (\form -> { form | name = string }) model
        --
        --
        --EnteredItemLogDescription string ->
        --    itemLogUpdateForm (\form -> { form | description = string }) model
        --
        --
        --SelectedItemLogUser string ->
        --    itemLogUpdateForm (\form -> { form | user = string }) model
        --
        --
        --SelectedItemLogItem string ->
        --    itemLogUpdateForm (\form -> { form | item = string }) model
        ViewItem itemId ->
            ( { model | selectedItemId = itemId }, Cmd.none )

        --ViewItemLog itemLogId ->
        --   ( { model | selectedItemLogId = itemLogId }, Cmd.none )
        SubmittedItemForm ->
            case ItemEdit.itemValidate model.itemForm of
                Ok validForm ->
                    ( { model | problems = [], loading = Loading.On }
                    , if model.item.id > 0 then
                        itemUpdate model validForm

                      else
                        itemAdd model validForm
                    )

                Err problems ->
                    ( { model | problems = problems, loading = Loading.Off }
                    , Cmd.none
                    )

        --SubmittedItemLogForm ->
        AddedItem result ->
            case result of
                Ok res ->
                    ( { model | apiActionResponse = res, loading = Loading.Off, itemForm = emptyItemForm }, loadItemById model res.resourceId )

                Err error ->
                    let
                        serverErrors =
                            decodeErrors error
                                |> List.map ServerError
                    in
                    ( { model | problems = List.append model.problems serverErrors, loading = Loading.Off, session = sessionGivenAuthError error model }
                    , Cmd.none
                    )

        --AddedItemLog result ->
        LoadedItem (Err error) ->
            ( { model
                | item = emptyItem
                , itemForm = emptyItemForm
                , loading = Loading.Off
                , session = sessionGivenAuthError error model
              }
            , Cmd.none
            )

        LoadedItem (Ok res) ->
            let
                itemForm =
                    { id = res.id
                    , silverNumber = String.fromInt res.silverNumber
                    , userId = res.userId
                    , viewPermissions = res.viewPermissions
                    , name = res.name
                    , itemType = res.itemType
                    , description = res.description
                    , latitude = String.fromFloat (toFloat res.latitudeI / 10000000)
                    , longitude = String.fromFloat (toFloat res.longitudeI / 10000000)
                    , altitude = String.fromFloat res.altitude
                    , status = res.status
                    , ageGroup = res.ageGroup
                    , height = String.fromFloat res.height
                    , latinName = res.latinName
                    , diameterAt = String.fromFloat res.diameterAt
                    , renderHints = res.renderHints
                    , fruitType = res.fruitType
                    , croppingSeason = res.croppingSeason
                    , fruitStorage = res.fruitStorage
                    , pollinatingGroup = res.pollinatingGroup
                    }
            in
            ( { model | item = res, itemForm = itemForm, loading = Loading.Off }
            , Cmd.none
            )

        FilterByRegion ->
            ( { model | page = ItemList, loading = Loading.Off }, Cmd.batch [ loadItemsForRegion model, loadRegions model ] )

        GotRegionPlotElement (Err _) ->
            ( model, Cmd.none )

        GotRegionPlotElement (Ok element) ->
            ( { model | regionPlotWidth = element.scene.width }, Cmd.none )


sessionGivenAuthError : Http.Error -> Model -> Session
sessionGivenAuthError error model =
    if error == BadStatus 401 then
        emptySession

    else
        model.session


decodeErrors : Http.Error -> List String
decodeErrors error =
    case error of
        Timeout ->
            [ "Timeout exceeded" ]

        NetworkError ->
            [ "Network error" ]

        BadBody body ->
            [ body ]

        BadUrl url ->
            [ "Malformed url: " ++ url ]

        BadStatus 401 ->
            [ "Invalid Username or Password" ]

        err ->
            [ "Server error" ]


fromPair : ( String, List String ) -> List String
fromPair ( field, errors ) =
    List.map (\error -> field ++ " " ++ error) errors


urlForPage : Page -> String
urlForPage page =
    case page of
        Profile ->
            "/profile"

        Home ->
            "/"

        Login ->
            "/login"

        Logout ->
            "/logout"

        Register _ _ ->
            "/register"

        Concepts string ->
            "/concepts/" ++ string

        ConceptsEdit string ->
            "/concepts/" ++ string ++ "/edit"

        ConceptsList ->
            "/concepts"

        NotFound ->
            ""

        AddConcept ->
            "/add_concept"

        ItemList ->
            "/items"

        AddItem ->
            "/add_item"

        ItemEdit string ->
            "/items/" ++ string ++ "/edit"

        ItemLogList int ->
            "/items/" ++ String.fromInt int ++ "/logs"

        UsersList ->
            "/users"

        UsersEdit string ->
            "/users/" ++ string ++ "/edit"



--AddItemLog string ->
--    "/items/" ++ string ++ "/add_log"
--ItemLogEdit item_string log_string ->
--    "/items/" ++ item_string ++ "/logs/" ++ log_string ++ "/edit"


urlUpdate : Url -> Model -> ( Model, Cmd Msg )
urlUpdate url model =
    case UrlParser.parse routeParser url of
        Nothing ->
            ( { model | page = NotFound }, Cmd.none )

        Just page ->
            ( case page of
                AddConcept ->
                    { model
                        | page = page
                        , concept = emptyConcept
                        , conceptForm = emptyConceptForm
                        , conceptTagForm = { tag = "" }
                    }

                --AddItemLog ->
                --    { model
                --        | page = page
                --        , itemLog = emptyItemLog
                --        , itemLogForm = emptyItemLogForm
                --    }
                AddItem ->
                    { model
                        | page = page
                        , item = emptyItem
                        , itemForm = emptyItemForm
                    }

                Register email verification ->
                    { model
                        | page = page
                        , registerForm =
                            { email = Maybe.withDefault "" email
                            , password = ""
                            , password_confirm = ""
                            , verification = Maybe.withDefault "" verification
                            }
                    }

                ItemList ->
                    { model
                        | page = page
                        , itemListFilter = ""
                    }

                _ ->
                    { model | page = page }
            , case page of
                Profile ->
                    Cmd.batch [ loadProfile model.session.loginToken ]

                Home ->
                    Cmd.batch [ loadConcept "index", loadConceptTagsList model, loadRegions model ]

                Concepts tag ->
                    Cmd.batch [ loadConcept tag ]

                ConceptsEdit id ->
                    let
                        conceptId =
                            Maybe.withDefault 0 (String.toInt id)
                    in
                    Cmd.batch [ loadConceptById conceptId, loadConceptTagsById conceptId ]

                ConceptsList ->
                    Cmd.batch [ loadConcepts model, loadConceptTagsList model ]

                Login ->
                    Cmd.none

                Logout ->
                    Cmd.none

                Register _ _ ->
                    Cmd.none

                AddConcept ->
                    Cmd.none

                ItemList ->
                    Cmd.batch [ loadItemsForRegion model, loadRegions model, Browser.Dom.getElement "region-plot" |> Task.attempt GotRegionPlotElement ]

                ItemLogList itemId ->
                    Cmd.batch [ loadItemById model itemId, loadItemLogs model itemId ]

                AddItem ->
                    Cmd.none

                --AddItemLog ->
                --    Cmd.none
                ItemEdit id ->
                    let
                        itemId =
                            Maybe.withDefault 0 (String.toInt id)
                    in
                    Cmd.batch [ loadItemById model itemId ]

                --ItemLogEdit id ->
                --    let
                --        itemLogId =
                --            Maybe.withDefault 0 (String.toInt id)
                --    in
                --    Cmd.batch [ loadItemLogById model itemLogId ]
                NotFound ->
                    Cmd.none

                UsersList ->
                    Cmd.batch [ loadUsers model ]

                UsersEdit string ->
                    let
                        userId =
                            Maybe.withDefault 0 (String.toInt string)
                    in
                    Cmd.batch [ loadUserById model userId ]
            )


routeParser : Parser (Page -> a) a
routeParser =
    UrlParser.oneOf
        [ UrlParser.map Home top
        , UrlParser.map Login (s "login")
        , UrlParser.map Logout (s "logout")
        , UrlParser.map Register (s "register" <?> Query.string "email" <?> Query.string "verification")
        , UrlParser.map Profile (s "profile")
        , UrlParser.map UsersEdit (s "users" </> string </> s "edit")
        , UrlParser.map UsersList (s "users")
        , UrlParser.map Concepts (s "concepts" </> string)
        , UrlParser.map ConceptsEdit (s "concepts" </> string </> s "edit")
        , UrlParser.map ConceptsList (s "concepts")
        , UrlParser.map AddConcept (s "add_concept")
        , UrlParser.map ItemList (s "items")
        , UrlParser.map AddItem (s "add_item")
        , UrlParser.map ItemEdit (s "items" </> string </> s "edit")
        , UrlParser.map ItemLogList (s "items" </> int </> s "logs")

        --, UrlParser.map AddItemLog (s "items" </> string </> s "add_log")
        --, UrlParser.map ItemLogEdit (s "items" </> string </> s "logs" </> string </> s "edit")
        ]



-- HTTP


loadUser : String -> Int -> Cmd Msg
loadUser token userId =
    Http.request
        { method = "GET"
        , url = "/api/users/" ++ String.fromInt userId
        , expect = Http.expectJson LoadedUser userDecoder
        , headers = [ authHeader token ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadUserById : Model -> Int -> Cmd Msg
loadUserById model userId =
    Http.request
        { method = "GET"
        , url = "/api/users/" ++ String.fromInt userId
        , expect = Http.expectJson LoadedOtherUser userDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadProfile : String -> Cmd Msg
loadProfile token =
    Http.request
        { method = "GET"
        , url = "/api/users/0"
        , expect = Http.expectJson LoadedProfile profileDecoder
        , headers = [ authHeader token ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadConcept : String -> Cmd Msg
loadConcept concept =
    Http.request
        { method = "GET"
        , url = "/api/concept/" ++ concept
        , expect = Http.expectJson LoadedConcept conceptDecoder
        , headers = []
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ case model.navState of
            Just navState ->
                Navbar.subscriptions navState NavMsg

            Nothing ->
                Sub.none
        , Time.every (60 * 1000) TimeTick
        ]
