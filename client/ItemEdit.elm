module ItemEdit exposing (..)

import Bootstrap.Button as Button
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Form as Form
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Select as Select
import FormValidation exposing (viewProblem)
import Html exposing (Html, div, h1, text, ul)
import Html.Attributes exposing (class, for)
import Html.Events exposing (onSubmit)
import Http
import ItemLogList exposing (itemCard)
import Json.Encode as Encode exposing (Value)
import Loading
import Types exposing (ApiActionResponse, ConceptForm, ConceptTag, ConceptTagForm, ItemForm, ItemTypeType(..), Model, Msg(..), Problem(..), Tag, UserPermissionsType(..), ValidatedField(..), apiActionDecoder, authHeader, formatBng, formatDate, formatLatLon, formatSilverNumber, itemTypeFromInt, itemTypeToInt, itemTypeToString, userPermissionsToInt)


pageItemEdit : Model -> List (Html Msg)
pageItemEdit model =
    [ if model.loading == Loading.On || model.loggedInUser.permissions == UserPermissionsNone || model.loggedInUser.permissions == UserPermissionsUser then
        Html.br [] []

      else
        div [ class "container page" ]
            [ div [ class "row" ]
                [ if model.loading == Loading.Off && model.item.id > 0 then
                    div [ class "col-12" ]
                        [ itemCard model.item
                        , h1 [ class "text-xs-center" ] [ text "Edit Item" ]
                        , viewItemForm model
                        ]

                  else if model.loading == Loading.Off then
                    div [ class "col-12" ]
                        [ h1 [ class "text-xs-center" ] [ text "Loading Item Failed" ]
                        ]

                  else
                    div [ class "col-12" ] [ h1 [ class "text-xs-center" ] [ text "Loading Item" ] ]
                ]
            ]
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    ]


pageAddItem : Model -> List (Html Msg)
pageAddItem model =
    [ if model.loggedInUser.permissions == UserPermissionsNone || model.loggedInUser.permissions == UserPermissionsUser then
        Html.br [] []

      else
        div [ class "container page" ]
            [ div [ class "row" ]
                [ div [ class "col-12" ]
                    [ h1 [ class "text-xs-center" ] [ text "Add Item" ]
                    , viewItemForm model
                    ]
                ]
            ]
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    ]


viewSelectableItemType : ItemForm -> Int -> Select.Item Msg
viewSelectableItemType form itemTypeIndex =
    let
        itemType =
            itemTypeFromInt itemTypeIndex
    in
    Select.item [ Html.Attributes.selected (form.itemType == itemType), Html.Attributes.value (String.fromInt (itemTypeToInt itemType)) ] [ text (itemTypeToString itemType) ]


viewItemForm : Model -> Html Msg
viewItemForm model =
    Form.form [ onSubmit SubmittedItemForm ]
        [ Form.group []
            [ Form.label [ for "silverNumber" ] [ text "Silver Number" ]
            , Input.number
                [ Input.id "silverNumber"
                , Input.placeholder "Silver Number"
                , Input.onInput EnteredItemSilverNumber
                , Input.value model.itemForm.silverNumber
                ]
            , Form.invalidFeedback [] [ text "Please enter a unique number for the item" ]
            ]
        , Form.group []
            [ Form.label [ for "viewPermissions" ] [ text "View Permissions" ]
            , Html.br [] []
            , ButtonGroup.radioButtonGroup []
                [ ButtonGroup.radioButton
                    (model.itemForm.viewPermissions == Types.UserPermissionsNone)
                    [ Button.primary, Button.onClick <| Types.SelectedViewPermissions Types.UserPermissionsNone ]
                    [ text "All" ]
                , ButtonGroup.radioButton
                    (model.itemForm.viewPermissions == Types.UserPermissionsUser)
                    [ Button.primary, Button.onClick <| Types.SelectedViewPermissions Types.UserPermissionsUser ]
                    [ text "User" ]
                , ButtonGroup.radioButton
                    (model.itemForm.viewPermissions == Types.UserPermissionsEditor)
                    [ Button.primary, Button.onClick <| Types.SelectedViewPermissions Types.UserPermissionsEditor ]
                    [ text "Editor" ]
                , ButtonGroup.radioButton
                    (model.itemForm.viewPermissions == Types.UserPermissionsAdmin)
                    [ Button.primary, Button.onClick <| Types.SelectedViewPermissions Types.UserPermissionsAdmin ]
                    [ text "Admin" ]
                ]
            , Form.invalidFeedback [] [ text "Please select the permissions required to view this item" ]
            ]
        , Form.group []
            [ Form.label [ for "name" ] [ text "Name" ]
            , Input.text
                [ Input.id "name"
                , Input.placeholder "Name"
                , Input.onInput EnteredItemName
                , Input.value model.itemForm.name
                ]
            , Form.invalidFeedback [] [ text "Please enter the name of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "itemType" ] [ text "Type" ]
            , Select.select
                [ Select.id "itemType"
                , Select.onChange SelectedItemType
                ]
                (List.concat
                    [ List.singleton (Select.item [ Html.Attributes.value "0" ] [ text "-[Select Type]-" ])
                    , List.map (viewSelectableItemType model.itemForm) (List.range 1 (itemTypeToInt ItemTypeMax - 1))
                    ]
                )
            ]
        , Form.group []
            [ Form.label [ for "description" ] [ text "Description" ]
            , Input.text
                [ Input.id "description"
                , Input.placeholder "Description"
                , Input.onInput EnteredItemDescription
                , Input.value model.itemForm.description
                ]
            , Form.invalidFeedback [] [ text "Please enter the description of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "latitude" ] [ text "Latitude" ]
            , Input.text
                [ Input.id "latitude"
                , Input.placeholder "Latitude"
                , Input.onInput EnteredItemLatitude
                , Input.value model.itemForm.latitude
                ]
            , Form.invalidFeedback [] [ text "Please enter the latitude of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "latitude" ] [ text "Longitude" ]
            , Input.text
                [ Input.id "longitude"
                , Input.placeholder "Longitude"
                , Input.onInput EnteredItemLongitude
                , Input.value model.itemForm.longitude
                ]
            , Form.invalidFeedback [] [ text "Please enter the longitude of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "altitude" ] [ text "Altitude" ]
            , Input.text
                [ Input.id "altitude"
                , Input.placeholder "Altitude"
                , Input.onInput EnteredItemAltitude
                , Input.value model.itemForm.altitude
                ]
            , Form.invalidFeedback [] [ text "Please enter the altitude of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "status" ] [ text "Status" ]
            , Input.text
                [ Input.id "status"
                , Input.placeholder "Status"
                , Input.onInput EnteredItemStatus
                , Input.value model.itemForm.status
                ]
            , Form.invalidFeedback [] [ text "Please enter the status of the item" ]
            ]
        , if model.itemForm.itemType == ItemTypeTree || model.itemForm.itemType == ItemTypeTreeFruit then
            Form.group []
                [ Form.label [ for "ageGroup" ] [ text "Age Group" ]
                , Input.text
                    [ Input.id "ageGroup"
                    , Input.placeholder "Age Group"
                    , Input.onInput EnteredItemAgeGroup
                    , Input.value model.itemForm.ageGroup
                    ]
                , Form.invalidFeedback [] [ text "Please enter the age group of the item" ]
                ]

          else
            Html.div [] []
        , Form.group []
            [ Form.label [ for "height" ] [ text "Height" ]
            , Input.text
                [ Input.id "height"
                , Input.placeholder "Height"
                , Input.onInput EnteredItemHeight
                , Input.value model.itemForm.height
                ]
            , Form.invalidFeedback [] [ text "Please enter the height of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "latinName" ] [ text "Latin Name" ]
            , Input.text
                [ Input.id "latinName"
                , Input.placeholder "Latin Name"
                , Input.onInput EnteredItemLatinName
                , Input.value model.itemForm.latinName
                ]
            , Form.invalidFeedback [] [ text "Please enter the latin name of the item" ]
            ]
        , Form.group []
            [ Form.label [ for "diameterAt" ] [ text "Diameter at 1m from the ground, in meters" ]
            , Input.text
                [ Input.id "diameterAt"
                , Input.placeholder "Diameter At"
                , Input.onInput EnteredItemDiameterAt
                , Input.value model.itemForm.diameterAt
                ]
            , Form.invalidFeedback [] [ text "Please enter the diameter of the item in meters at 1m off the ground" ]
            ]
        , Form.group []
            [ Form.label [ for "renderHints" ] [ text "Render Hints" ]
            , Input.text
                [ Input.id "renderHints"
                , Input.placeholder "Render Hints"
                , Input.onInput EnteredItemRenderHints
                , Input.value model.itemForm.renderHints
                ]
            , Form.invalidFeedback [] [ text "Please enter the render hints of the item" ]
            ]
        , if model.itemForm.itemType == ItemTypeTreeFruit then
            Form.group []
                [ Form.label [ for "fruitType" ] [ text "Fruit Type" ]
                , Input.text
                    [ Input.id "fruitType"
                    , Input.placeholder "Fruit Type"
                    , Input.onInput EnteredItemFruitType
                    , Input.value model.itemForm.fruitType
                    ]
                , Form.invalidFeedback [] [ text "Please enter the fruit type of the item" ]
                ]

          else
            Html.div [] []
        , if model.itemForm.itemType == ItemTypeTreeFruit then
            Form.group []
                [ Form.label [ for "croppingSeason" ] [ text "Cropping Season" ]
                , Input.text
                    [ Input.id "croppingSeason"
                    , Input.placeholder "Cropping Season"
                    , Input.onInput EnteredItemCroppingSeason
                    , Input.value model.itemForm.croppingSeason
                    ]
                , Form.invalidFeedback [] [ text "Please enter the cropping season of the item" ]
                ]

          else
            Html.div [] []
        , if model.itemForm.itemType == ItemTypeTreeFruit then
            Form.group []
                [ Form.label [ for "fruitStorage" ] [ text "Fruit Storage" ]
                , Input.text
                    [ Input.id "fruitStorage"
                    , Input.placeholder "Fruit Storage"
                    , Input.onInput EnteredItemFruitStorage
                    , Input.value model.itemForm.fruitStorage
                    ]
                , Form.invalidFeedback [] [ text "Please enter the fruit Storage of the item" ]
                ]

          else
            Html.div [] []
        , if model.itemForm.itemType == ItemTypeTreeFruit then
            Form.group []
                [ Form.label [ for "pollinatingGroup" ] [ text "Pollinating Group" ]
                , Input.text
                    [ Input.id "pollinatingGroup"
                    , Input.placeholder "Pollinating Group"
                    , Input.onInput EnteredItemPollinatingGroup
                    , Input.value model.itemForm.pollinatingGroup
                    ]
                , Form.invalidFeedback [] [ text "Please enter the pollinating group of the item" ]
                ]

          else
            Html.div [] []
        , ul [ class "error-messages" ]
            (List.map viewProblem model.problems)
        , Button.button [ Button.primary ]
            [ text "Save Item" ]
        , Loading.render Loading.DoubleBounce Loading.defaultConfig model.loading
        ]


itemFieldsToValidate : List ValidatedField
itemFieldsToValidate =
    [ Name
    , SilverNumber
    ]


type ItemTrimmedForm
    = ItemTrimmed ItemForm


itemTrimFields : ItemForm -> ItemTrimmedForm
itemTrimFields form =
    let
        silverNumberMaybe =
            String.toInt form.silverNumber

        silverNumberString =
            case silverNumberMaybe of
                Just silverNumber ->
                    String.fromInt silverNumber

                Nothing ->
                    ""
    in
    ItemTrimmed
        { id = form.id
        , silverNumber = silverNumberString
        , userId = form.userId
        , viewPermissions = form.viewPermissions
        , name = String.trim form.name
        , itemType = form.itemType
        , description = String.trim form.description
        , latitude = String.trim form.latitude
        , longitude = String.trim form.longitude
        , altitude = String.trim form.altitude
        , status = String.trim form.status
        , ageGroup = String.trim form.ageGroup
        , height = String.trim form.height
        , latinName = String.trim form.latinName
        , diameterAt = String.trim form.diameterAt
        , renderHints = String.trim form.renderHints
        , fruitType = String.trim form.fruitType
        , croppingSeason = String.trim form.croppingSeason
        , fruitStorage = String.trim form.fruitStorage
        , pollinatingGroup = String.trim form.pollinatingGroup
        }


validateField : ItemTrimmedForm -> ValidatedField -> List Problem
validateField (ItemTrimmed form) field =
    let
        silverNumberInt =
            Maybe.withDefault 0 (String.toInt form.silverNumber)
    in
    List.map (InvalidEntry field) <|
        case field of
            Name ->
                if form.name == "" then
                    [ "you must enter a name" ]

                else
                    []

            SilverNumber ->
                if silverNumberInt == 0 then
                    [ "you must create a unique number for each item" ]

                else
                    []

            _ ->
                []


itemValidate : ItemForm -> Result (List Problem) ItemTrimmedForm
itemValidate form =
    let
        trimmedForm =
            itemTrimFields form
    in
    case List.concatMap (validateField trimmedForm) itemFieldsToValidate of
        [] ->
            Ok trimmedForm

        problems ->
            Err problems


itemUpdateForm : (ItemForm -> ItemForm) -> Model -> ( Model, Cmd Msg )
itemUpdateForm transform model =
    ( { model | itemForm = transform model.itemForm }, Cmd.none )



-- HTTP


itemBody : Model -> ItemForm -> Http.Body
itemBody model form =
    let
        silverNumberInt =
            Maybe.withDefault 0 (String.toInt form.silverNumber)

        latitudeInt =
            case String.toFloat form.latitude of
                Just f ->
                    round (f * 10000000)

                Nothing ->
                    0

        longitudeInt =
            case String.toFloat form.longitude of
                Just f ->
                    round (f * 10000000)

                Nothing ->
                    0

        altitude =
            case String.toFloat form.altitude of
                Just i ->
                    i

                Nothing ->
                    0

        height =
            case String.toFloat form.height of
                Just f ->
                    f

                Nothing ->
                    0.0

        diameterAt =
            case String.toFloat form.diameterAt of
                Just f ->
                    f

                Nothing ->
                    0.0

        viewPermissions =
            userPermissionsToInt form.viewPermissions

        itemType =
            itemTypeToInt form.itemType
    in
    Encode.object
        [ ( "Id", Encode.int model.item.id )
        , ( "SilverNumber", Encode.int silverNumberInt )
        , ( "UserId", Encode.int form.userId )
        , ( "ViewPermissions", Encode.int viewPermissions )
        , ( "Name", Encode.string form.name )
        , ( "ItemType", Encode.int itemType )
        , ( "Description", Encode.string form.description )
        , ( "LatitudeI", Encode.int latitudeInt )
        , ( "LongitudeI", Encode.int longitudeInt )
        , ( "Altitude", Encode.float altitude )
        , ( "Status", Encode.string form.status )
        , ( "AgeGroup", Encode.string form.ageGroup )
        , ( "Height", Encode.float height )
        , ( "LatinName", Encode.string form.latinName )
        , ( "DiameterAt", Encode.float diameterAt )
        , ( "RenderHints", Encode.string form.renderHints )
        , ( "FruitType", Encode.string form.fruitType )
        , ( "CroppingSeason", Encode.string form.croppingSeason )
        , ( "FruitStorage", Encode.string form.fruitStorage )
        , ( "PollinatingGroup", Encode.string form.pollinatingGroup )
        ]
        |> Http.jsonBody


itemUpdate : Model -> ItemTrimmedForm -> Cmd Msg
itemUpdate model (ItemTrimmed form) =
    let
        body =
            itemBody model form
    in
    Http.request
        { method = "PUT"
        , url = "/api/items/" ++ String.fromInt model.item.id
        , expect = Http.expectJson AddedItem apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }


itemAdd : Model -> ItemTrimmedForm -> Cmd Msg
itemAdd model (ItemTrimmed form) =
    let
        body =
            itemBody model form
    in
    Http.request
        { method = "POST"
        , url = "/api/items"
        , expect = Http.expectJson AddedItem apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }
