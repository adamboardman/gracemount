module ItemList exposing (..)

import Bootstrap.Button as Button
import Bootstrap.Form.Select as Select
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Dict
import FormValidation exposing (viewProblem)
import Html exposing (Html, a, div, h4, input, text)
import Html.Attributes exposing (class, href, id, placeholder, value)
import Html.Events exposing (..)
import Http exposing (emptyBody)
import ItemLogList exposing (itemCard)
import Json.Decode exposing (Decoder, list)
import List.Extra
import Loading
import Login exposing (userIsEditor)
import String exposing (contains)
import Svg exposing (Svg, circle, svg)
import Svg.Attributes exposing (cx, cy, fill, height, r, stroke, strokeWidth, viewBox, width)
import Types exposing (..)


pageItemList : Model -> List (Html Msg)
pageItemList model =
    let
        itemList =
            if model.searchRegionId == 0 then
                model.itemList

            else
                Maybe.withDefault [] (Dict.get model.searchRegionId model.itemListForRegion)

        filteredItemList =
            if String.length model.itemListFilter > 0 then
                List.filterMap (removeUnmatched model.itemListFilter) itemList

            else
                itemList
    in
    [ h4 [] [ text "Items" ]
    , if model.loading == Loading.Off && model.searchRegionId > 0 then
        div [ id "region-plot" ] (regionPlot model itemList filteredItemList)

      else
        div [] []
    , div [] (filterOptions model)
    , if model.loading == Loading.Off && List.length filteredItemList > 1 then
        div [] (List.map (itemSummary model) filteredItemList)

      else if model.loading == Loading.Off && List.length filteredItemList > 0 then
        itemCard (Maybe.withDefault emptyItem (List.head filteredItemList))

      else
        div [] []
    , div [] (List.map viewProblem model.problems)
    ]


osgbNorthing : Item -> Float
osgbNorthing item =
    item.osgb.lat


osgbEasting : Item -> Float
osgbEasting item =
    item.osgb.lng


itemMatchesId : Item -> Item -> Bool
itemMatchesId item1 item2 =
    item1.id == item2.id


itemCircle : List Item -> (Float -> Float) -> (Float -> Float) -> Float -> Item -> Svg msg
itemCircle filteredItems plotterX plotterY height item =
    let
        strokeColour =
            case List.Extra.find (itemMatchesId item) filteredItems of
                Just _ ->
                    "black"

                Nothing ->
                    "lightgrey"

        fillColour =
            case item.itemType of
                ItemTypeAllOrUnknown ->
                    "yellow"

                ItemTypeTree ->
                    "darkgreen"

                ItemTypeTreeFruit ->
                    "forestgreen"

                ItemTypeInfrastructure ->
                    "grey"

                ItemTypeSmokeDetector ->
                    "red"

                ItemTypeMax ->
                    "red"
    in
    circle
        [ cx (String.fromFloat (plotterX item.osgb.lng))
        , cy (String.fromFloat (height - plotterY item.osgb.lat))
        , r "5"
        , fill fillColour
        , stroke strokeColour
        , strokeWidth "2"
        ]
        []


plotter : Float -> Float -> Float -> Float -> Float
plotter min multiplier margin pos =
    margin + (pos - min) * multiplier


regionPlot : Model -> List Item -> List Item -> List (Html Msg)
regionPlot model itemList filteredItems =
    let
        maxX =
            Maybe.withDefault 0 (List.maximum (List.map osgbEasting itemList))

        minX =
            Maybe.withDefault 0 (List.minimum (List.map osgbEasting itemList))

        maxY =
            Maybe.withDefault 0 (List.maximum (List.map osgbNorthing itemList))

        minY =
            Maybe.withDefault 0 (List.minimum (List.map osgbNorthing itemList))

        diffX =
            maxX - minX

        diffY =
            maxY - minY

        margin =
            7

        w =
            model.regionPlotWidth

        scaleX =
            (w - margin * 2) / diffX

        h =
            diffY * scaleX

        scaleY =
            (h - margin * 2) / diffY

        plotterX =
            plotter minX scaleX margin

        plotterY =
            plotter minY scaleY margin
    in
    [ svg
        [ viewBox ("0 0 " ++ String.fromFloat w ++ " " ++ String.fromFloat h)
        , width "100%"
        ]
        (List.map (itemCircle filteredItems plotterX plotterY h) itemList)
    ]


filterOptions : Model -> List (Html Msg)
filterOptions model =
    [ div [ class "input-group" ]
        [ case model.regionList of
            Just regionList ->
                Select.select
                    [ Select.id "region"
                    , Select.onChange SelectedRegion
                    ]
                    (List.concat
                        [ List.singleton (Select.item [ Html.Attributes.value "0" ] [ text "-[Any Site]-" ])
                        , List.map (viewSelectableRegion model.searchRegionId) regionList
                        ]
                    )

            Nothing ->
                div [] []
        , input [ class "form-control", placeholder "Filter", value model.itemListFilter, onInput EnteredFilter ] []

        --, Button.button [ Button.outlineSecondary, Button.onClick FilterByRegion ] [ text "Search" ]
        ]
    ]


filterOptionsGrid : Model -> List (Html Msg)
filterOptionsGrid model =
    [ Grid.row [ Row.middleMd ]
        [ Grid.col [ Col.md1 ]
            [ case model.regionList of
                Just regionList ->
                    Select.select
                        [ Select.id "region"
                        , Select.onChange SelectedRegion
                        ]
                        (List.concat
                            [ List.singleton (Select.item [ Html.Attributes.value "0" ] [ text "- [Any Location] -" ])
                            , List.map (viewSelectableRegion model.searchRegionId) regionList
                            ]
                        )

                Nothing ->
                    div [] []
            ]
        , Grid.col [ Col.md2 ] [ Button.button [ Button.outlineSecondary, Button.onClick FilterByRegion ] [ text "Search" ] ]
        , Grid.col [ Col.md3 ] [ input [ class "filter-box", placeholder "Filter", value model.itemListFilter, onInput EnteredFilter ] [] ]
        ]
    ]


viewSelectableRegion : Int -> Region -> Select.Item Msg
viewSelectableRegion regionId region =
    Select.item [ Html.Attributes.selected (regionId == region.id), Html.Attributes.value (String.fromInt region.id) ] [ text region.name ]


removeUnmatched : String -> Item -> Maybe Item
removeUnmatched filter item =
    let
        itemName =
            String.toLower item.name

        stringFilter =
            String.toLower filter

        intFilter =
            Maybe.withDefault 0 (String.toInt filter)
    in
    if item.silverNumber == intFilter then
        Just item

    else if contains stringFilter itemName then
        Just item

    else
        Nothing


itemSummary : Model -> Item -> Html Msg
itemSummary model item =
    div []
        [ text (" Silver Number: " ++ formatSilverNumber item.silverNumber)
        , text (", Name: " ++ item.name ++ " ")
        , if String.length model.item.description > 0 then
            text (", Description: " ++ item.description)

          else
            text ""
        , text " "
        , a [ href ("/items/" ++ String.fromInt item.id ++ "/logs") ] [ text "(view)" ]
        , text " "
        , if userIsEditor model then
            text ("Created by UserId: " ++ String.fromInt item.userId)

          else
            text ""
        , text " "
        , if userIsEditor model then
            a [ href ("/items/" ++ String.fromInt item.id ++ "/edit") ] [ text "(edit)" ]

          else
            text ""
        ]



-- HTTP


loadItems : Model -> Cmd Msg
loadItems model =
    Http.request
        { method = "GET"
        , url = "/api/items"
        , expect = Http.expectJson LoadedItems itemListDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadItemsForRegion : Model -> Cmd Msg
loadItemsForRegion model =
    Http.request
        { method = "GET"
        , url = "/api/items_for_region/" ++ String.fromInt model.searchRegionId
        , expect = Http.expectJson LoadedItemsForRegion itemListDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadItemById : Model -> Int -> Cmd Msg
loadItemById model itemId =
    Http.request
        { method = "GET"
        , url = "/api/items/" ++ String.fromInt itemId
        , expect = Http.expectJson LoadedItem itemDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadRegions : Model -> Cmd Msg
loadRegions model =
    if model.regionList == Nothing then
        Http.request
            { method = "GET"
            , url = "/api/regions"
            , expect = Http.expectJson LoadedRegions regionListDecoder
            , headers = [ authHeader model.session.loginToken ]
            , body = emptyBody
            , timeout = Nothing
            , tracker = Nothing
            }

    else
        Cmd.none


itemListDecoder : Decoder (List Item)
itemListDecoder =
    list itemDecoder


regionListDecoder : Decoder (List Region)
regionListDecoder =
    list regionDecoder
