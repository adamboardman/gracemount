module ConceptsEdit exposing (..)

import Bootstrap.Button as Button
import Bootstrap.Form as Form
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.Textarea as Textarea
import Bootstrap.Modal as Modal
import FormValidation exposing (viewProblem)
import Html exposing (Html, div, h1, text, ul)
import Html.Attributes exposing (class, for)
import Html.Events exposing (onSubmit)
import Http exposing (emptyBody)
import Json.Encode as Encode exposing (Value)
import Loading
import Markdown
import Set exposing (Set)
import Types exposing (ApiActionResponse, ConceptForm, ConceptTag, ConceptTagForm, Model, Msg(..), Problem(..), Tag, ValidatedField(..), apiActionDecoder, authHeader, conceptDecoder, conceptTagsListDecoder)


pageConceptsEdit : Model -> List (Html Msg)
pageConceptsEdit model =
    [ div [ class "container page" ]
        [ div [ class "row" ]
            [ if model.loading == Loading.Off && model.concept.id > 0 then
                div [ class "col-12" ]
                    [ h1 [ class "text-xs-center" ] [ text "Existing Concept" ]
                    , div [] <| Markdown.toHtml Nothing model.concept.full
                    , h1 [ class "text-xs-center" ] [ text "Edit Concept" ]
                    , viewConceptForm model
                    , viewConceptModal model
                    ]

              else if model.loading == Loading.Off then
                div [ class "col-12" ]
                    [ h1 [ class "text-xs-center" ] [ text "Loading Concept Failed" ]
                    ]

              else
                div [ class "col-12" ] [ h1 [ class "text-xs-center" ] [ text "Loading Concept" ] ]
            ]
        ]
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    ]


pageAddConcept : Model -> List (Html Msg)
pageAddConcept model =
    [ div [ class "container page" ]
        [ div [ class "row" ]
            [ div [ class "col-12" ]
                [ h1 [ class "text-xs-center" ] [ text "Add Concept" ]
                , viewConceptForm model
                , viewConceptModal model
                ]
            ]
        ]
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    , Html.br [] []
    ]


viewConceptForm : Model -> Html Msg
viewConceptForm model =
    Form.form [ onSubmit SubmittedConceptForm ]
        [ Form.group []
            [ Form.label [ for "name" ] [ text "Concept Name" ]
            , Input.text
                [ Input.id "name"
                , Input.placeholder "Concept Name"
                , Input.onInput EnteredConceptName
                , Input.value model.conceptForm.name
                ]
            , Form.invalidFeedback [] [ text "Please enter a name" ]
            ]
        , Form.group []
            [ Form.label [] [ text "Tags" ]
            , Form.row []
                (List.map (viewConceptFormTag model.conceptForm) model.conceptForm.tags)
            ]
        , if model.concept.id > 0 then
            Form.group []
                [ Button.button [ Button.secondary, Button.onClick ButtonConceptAddTag ] [ text "Add Tag" ]
                , Html.span [] [ text "\u{00A0}\u{00A0}" ]
                , Button.button [ Button.secondary, Button.disabled (Set.isEmpty model.conceptForm.tagsToDelete), Button.onClick ButtonConceptDeleteSelectedTags ]
                    [ text "Delete Selected Tags" ]
                ]

          else
            Form.group []
                [ Form.label [] [ text "Must save concept before adding tags" ]
                ]
        , Form.group []
            [ Form.label [ for "summary" ] [ text "Summary" ]
            , Textarea.textarea
                [ Textarea.id "summary"
                , Textarea.rows 3
                , Textarea.onInput EnteredConceptSummary
                , Textarea.value model.conceptForm.summary
                ]
            , Form.invalidFeedback [] [ text "Summary of concept for display with links" ]
            ]
        , Form.group []
            [ Form.label [ for "full" ] [ text "Full" ]
            , Textarea.textarea
                [ Textarea.id "full"
                , Textarea.rows 5
                , Textarea.onInput EnteredConceptFull
                , Textarea.value model.conceptForm.full
                ]
            , Form.invalidFeedback [] [ text "Full description of concept" ]
            ]
        , ul [ class "error-messages" ]
            (List.map viewProblem model.problems)
        , Button.button [ Button.primary ]
            [ text "Save Concept" ]
        , Loading.render Loading.DoubleBounce Loading.defaultConfig model.loading
        ]


viewConceptModal : Model -> Html Msg
viewConceptModal model =
    Modal.config CloseConceptAddTagModal
        |> Modal.small
        |> Modal.h5 [] [ text "Add Concept Tag" ]
        |> Modal.body []
            [ Form.form [ onSubmit SubmittedAddConceptTagForm ]
                [ Form.group
                    []
                    [ Form.label [ for "tagTag" ] [ text "Tag" ]
                    , Input.text
                        [ Input.id "tagTag"
                        , Input.placeholder "Enter tag"
                        , Input.onInput EnteredAddConceptTag
                        , Input.value model.conceptTagForm.tag
                        ]
                    , Form.invalidFeedback [] [ text "Please enter a tag" ]
                    ]
                ]
            , Button.button [ Button.primary, Button.onClick SubmittedAddConceptTagForm ]
                [ text "Add tag" ]
            , Loading.render Loading.DoubleBounce Loading.defaultConfig model.loading
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.secondary
                , Button.onClick CloseConceptAddTagModal
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.conceptShowTagModel


viewConceptFormTag : ConceptForm -> ConceptTag -> Form.Col Msg
viewConceptFormTag form tag =
    Form.col []
        [ Form.label [ for (String.fromInt tag.id) ]
            [ Checkbox.checkbox
                [ Checkbox.id (String.fromInt tag.id)
                , Checkbox.checked (tagsSetContains form.tagsToDelete tag.id)
                , Checkbox.onCheck (EnteredConceptTagCheckToDelete tag.id)
                ]
                tag.tag
            ]
        ]


tagsSetContains : Set Int -> Int -> Bool
tagsSetContains tags tag =
    Set.member tag tags


tagIsNotIn : Set Int -> ConceptTag -> Bool
tagIsNotIn tagIds tag =
    not (Set.member tag.id tagIds)


conceptFieldsToValidate : List ValidatedField
conceptFieldsToValidate =
    [ Name ]


conceptTagFieldsToValidate : List ValidatedField
conceptTagFieldsToValidate =
    [ TagTag ]


conceptUpdateForm : (ConceptForm -> ConceptForm) -> Model -> ( Model, Cmd Msg )
conceptUpdateForm transform model =
    ( { model | conceptForm = transform model.conceptForm }, Cmd.none )


conceptTagUpdateForm : (ConceptTagForm -> ConceptTagForm) -> Model -> ( Model, Cmd Msg )
conceptTagUpdateForm transform model =
    ( { model | conceptTagForm = transform model.conceptTagForm }, Cmd.none )


conceptValidate : ConceptForm -> Result (List Problem) ConceptTrimmedForm
conceptValidate form =
    let
        trimmedForm =
            conceptTrimFields form
    in
    case List.concatMap (validateField trimmedForm) conceptFieldsToValidate of
        [] ->
            Ok trimmedForm

        problems ->
            Err problems


conceptTagValidate : ConceptTagForm -> Result (List Problem) ConceptTagTrimmedForm
conceptTagValidate form =
    let
        trimmedForm =
            conceptTagTrimFields form
    in
    case List.concatMap (validateTagField trimmedForm) conceptTagFieldsToValidate of
        [] ->
            Ok trimmedForm

        problems ->
            Err problems


validateField : ConceptTrimmedForm -> ValidatedField -> List Problem
validateField (ConceptTrimmed form) field =
    List.map (InvalidEntry field) <|
        case field of
            Name ->
                if String.isEmpty form.name then
                    [ "name can't be blank." ]

                else
                    []

            _ ->
                []


validateTagField : ConceptTagTrimmedForm -> ValidatedField -> List Problem
validateTagField (ConceptTagTrimmed form) field =
    List.map (InvalidEntry field) <|
        case field of
            TagTag ->
                if String.isEmpty form.tag then
                    [ "tag can't be blank." ]

                else
                    []

            _ ->
                []


type ConceptTrimmedForm
    = ConceptTrimmed ConceptForm


conceptTrimFields : ConceptForm -> ConceptTrimmedForm
conceptTrimFields form =
    ConceptTrimmed
        { name = String.trim form.name
        , tags = form.tags
        , tagsToDelete = form.tagsToDelete
        , summary = String.trim form.summary
        , full = String.trim form.full
        }


type ConceptTagTrimmedForm
    = ConceptTagTrimmed ConceptTagForm


conceptTagTrimFields : ConceptTagForm -> ConceptTagTrimmedForm
conceptTagTrimFields form =
    ConceptTagTrimmed
        { tag = String.trim form.tag
        }



-- HTTP


conceptUpdate : Model -> ConceptTrimmedForm -> Cmd Msg
conceptUpdate model (ConceptTrimmed form) =
    let
        body =
            Encode.object
                [ ( "Id", Encode.int model.concept.id )
                , ( "Name", Encode.string form.name )
                , ( "Summary", Encode.string form.summary )
                , ( "Full", Encode.string form.full )
                ]
                |> Http.jsonBody
    in
    Http.request
        { method = "PUT"
        , url = "/api/concepts/" ++ String.fromInt model.concept.id
        , expect = Http.expectJson AddedConcept apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }


conceptAdd : Model -> ConceptTrimmedForm -> Cmd Msg
conceptAdd model (ConceptTrimmed form) =
    let
        body =
            Encode.object
                [ ( "Id", Encode.int model.concept.id )
                , ( "Name", Encode.string form.name )
                , ( "Summary", Encode.string form.summary )
                , ( "Full", Encode.string form.full )
                ]
                |> Http.jsonBody
    in
    Http.request
        { method = "POST"
        , url = "/api/concepts"
        , expect = Http.expectJson AddedConcept apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }


conceptTag : Model -> ConceptTagTrimmedForm -> Cmd Msg
conceptTag model (ConceptTagTrimmed form) =
    let
        body =
            Encode.object
                [ ( "ConceptId", Encode.int model.concept.id )
                , ( "Tag", Encode.string form.tag )
                ]
                |> Http.jsonBody
    in
    Http.request
        { method = "POST"
        , url = "/api/concept_tags"
        , expect = Http.expectJson (AddedConceptTag model.concept.id form.tag) apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }


loadConceptById : Int -> Cmd Msg
loadConceptById conceptId =
    Http.request
        { method = "GET"
        , url = "/api/concepts/" ++ String.fromInt conceptId
        , expect = Http.expectJson LoadedConcept conceptDecoder
        , headers = []
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


loadConceptTagsById : Int -> Cmd Msg
loadConceptTagsById conceptId =
    Http.request
        { method = "GET"
        , url = "/api/concepts/" ++ String.fromInt conceptId ++ "/tags"
        , expect = Http.expectJson LoadedConceptTags conceptTagsListDecoder
        , headers = []
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }


conceptDeleteSelectedTags : Model -> Cmd Msg
conceptDeleteSelectedTags model =
    let
        body =
            Encode.list Encode.int (Set.toList model.conceptForm.tagsToDelete) |> Http.jsonBody
    in
    Http.request
        { method = "DELETE"
        , url = "/api/concept_tags"
        , expect = Http.expectJson ConceptTagDeleted apiActionDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = body
        , timeout = Nothing
        , tracker = Nothing
        }
