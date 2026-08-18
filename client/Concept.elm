module Concept exposing (pageConcept)

import FormValidation exposing (viewProblem)
import Html exposing (Html, div, h4, text)
import Html.Attributes
import ItemList exposing (filterOptions)
import Markdown
import Types exposing (Concept, ConceptTag, DisplayableTag, Model, Msg(..), Tag)


pageConcept : Model -> List (Html Msg)
pageConcept model =
    [ h4 [] [ text model.concept.name ]
    , div [] <| Markdown.toHtml Nothing model.concept.full
    , case displayableTagsListMatchesStringInt model.displayableTagsList "index" of
        Just indexId ->
            if model.concept.id == indexId then
                div [] (filterOptions model)

            else
                div [] []

        Nothing ->
            div [] []
    , div [] (List.map viewProblem model.problems)
    , Html.img
        [ Html.Attributes.src "public/no-internet.svg"
        , Html.Attributes.height 0
        , Html.Attributes.width 0
        ]
        []
    ]


stringsMatch : String -> String -> Bool
stringsMatch string1 string2 =
    string1 == string2


displayableTagIdMatchesString : String -> DisplayableTag -> Maybe Int
displayableTagIdMatchesString string tag =
    if List.length (List.filter (stringsMatch string) tag.tags) > 0 then
        Just tag.id

    else
        Nothing


displayableTagsListMatchesStringInt : List DisplayableTag -> String -> Maybe Int
displayableTagsListMatchesStringInt tags string =
    let
        filtered =
            List.filterMap (displayableTagIdMatchesString string) tags
    in
    List.head filtered
