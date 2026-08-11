module ItemLogList exposing (..)

import Bootstrap.Card as Card
import Bootstrap.Card.Block as Block
import FormValidation exposing (viewProblem)
import GeoCoord exposing (latLonToBng)
import Html exposing (Html, div, h4, text)
import Html.Attributes exposing (class)
import Http exposing (emptyBody)
import Json.Decode exposing (Decoder, list)
import Login exposing (userIsEditor)
import Types exposing (..)


pageItemLogList : Model -> Int -> List (Html Msg)
pageItemLogList model id =
    [ itemCard model.item
    , Html.div [] []
    , h4 [] [ text "Item Logs" ]
    , if List.length model.itemLogList > 0 then
        div [] (List.map (itemLogSummary model) model.itemLogList)

      else
        Html.div [] [ text "(Nothing logged yet)" ]
    , div [] (List.map viewProblem model.problems)
    ]


itemCard : Item -> Html Msg
itemCard item =
    let
        latLon =
            { lon = toFloat item.longitudeI / 10000000, lat = toFloat item.latitudeI / 10000000 }

        bng =
            latLonToBng latLon
    in
    Card.config [ Card.outlineInfo ]
        |> Card.headerH3 []
            [ text (formatSilverNumber item.silverNumber)
            , text (" : " ++ item.name)
            , if String.length item.latinName > 0 then
                Html.span [] [ text " : ", Html.span [ class "text-muted" ] [ text item.latinName ] ]

              else
                text ""
            ]
        |> Card.block []
            [ if String.length item.description > 0 then
                Block.text [] [ text item.description ]

              else
                Block.text [] []
            , if String.length item.status > 0 then
                Block.text [] [ text item.status ]

              else
                Block.text [] []
            , if String.length item.ageGroup > 0 then
                Block.text [] [ text item.ageGroup ]

              else
                Block.text [] []
            , Block.text []
                [ if item.diameterAt > 0 then
                    text (formatDiameterAt item.diameterAt ++ ", ")

                  else
                    text ""
                , if item.height > 0 then
                    text (formatHeight item.height)

                  else
                    text ""
                ]
            , Block.text []
                [ if String.length item.fruitType > 0 then
                    text ("Fruit Type: " ++ item.fruitType ++ ", ")

                  else
                    text ""
                , if String.length item.croppingSeason > 0 then
                    text ("Cropping Season: " ++ item.croppingSeason ++ ", ")

                  else
                    text ""
                , if String.length item.fruitStorage > 0 then
                    text ("Fruit Storage: " ++ item.fruitStorage ++ ", ")

                  else
                    text ""
                , if String.length item.pollinatingGroup > 0 then
                    text ("Pollinating Group: " ++ item.pollinatingGroup)

                  else
                    text ""
                ]
            ]
        |> Card.footer []
            [ if not (item.latitudeI == 0 && item.longitudeI == 0) then
                Html.small [ class "text-muted" ]
                    [ text (formatLatLon latLon)
                    , text " | "
                    , text (formatBng bng)
                    , if item.altitude > 0 then
                        text (" | " ++ formatAltitude item.altitude)

                      else
                        text ""
                    ]

              else
                Html.div [] []
            ]
        |> Card.view


itemLogSummary : Model -> ItemLog -> Html Msg
itemLogSummary model itemLog =
    div []
        [ if userIsEditor model then
            text ("Created by UserId: " ++ String.fromInt itemLog.userId)

          else
            text " "
        , text (" Item Id: " ++ String.fromInt itemLog.itemId)
        , text (", Date: " ++ formatDateTime model itemLog.date)
        , text (", Name: " ++ itemLog.name)
        , text (", Description: " ++ itemLog.description)
        , text " "
        ]



-- HTTP


loadItemLogs : Model -> Int -> Cmd Msg
loadItemLogs model itemId =
    Http.request
        { method = "GET"
        , url = "/api/items/" ++ String.fromInt itemId ++ "/logs"
        , expect = Http.expectJson LoadedItemLogs itemLogListDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


itemLogListDecoder : Decoder (List ItemLog)
itemLogListDecoder =
    list itemLogDecoder
