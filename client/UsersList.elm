module UsersList exposing (..)

import Bootstrap.Table as Table exposing (Row, rowAttr)
import FormValidation exposing (viewProblem)
import Html exposing (Html, a, h4, text)
import Html.Attributes exposing (href, style)
import Http exposing (emptyBody)
import Json.Decode exposing (Decoder, list)
import Loading
import Types exposing (ApiActionResponse, Concept, ConceptTag, Model, Msg(..), Page(..), Problem(..), User, UserPermissionsType(..), ValidatedField(..), authHeader, userDecoder, userPermissionsToString)


userSummary : Model -> User -> Row Msg
userSummary model user =
    Table.tr
        [ if user.permissions == UserPermissionsNone then
            rowAttr (style "color" "grey")

          else
            rowAttr (style "" "")
        ]
        [ Table.td [] [ text user.firstName ]
        , Table.td [] [ text user.midNames ]
        , Table.td [] [ text user.lastName ]
        , Table.td [] [ text user.location ]
        , Table.td [] [ text user.email ]
        , Table.td [] [ text user.mobile ]
        , Table.td [] [ text (userPermissionsToString user.permissions) ]
        , Table.td [] [ a [ href ("/users/" ++ String.fromInt user.id ++ "/edit") ] [ text "(edit)" ] ]
        ]


pageUsersList : Model -> List (Html Msg)
pageUsersList model =
    List.concat
        [ [ h4 [] [ text "Users" ]
          , Table.table
                { options = [ Table.striped, Table.hover ]
                , thead =
                    Table.simpleThead
                        [ Table.th [] [ text "First" ]
                        , Table.th [] [ text "Mid" ]
                        , Table.th [] [ text "Last" ]
                        , Table.th [] [ text "Location" ]
                        , Table.th [] [ text "Email" ]
                        , Table.th [] [ text "Mobile" ]
                        , Table.th [] [ text "Permissions" ]
                        , Table.th [] [ text "Edit" ]
                        ]
                , tbody =
                    if model.loading == Loading.Off then
                        Table.tbody []
                            (List.map
                                (userSummary model)
                                model.usersList
                            )

                    else
                        Table.tbody [] []
                }
          , Html.br [] []
          , Html.div [] (List.map viewProblem model.problems)
          ]
        , [ Html.br [] []
          , Html.br [] []
          , Html.br [] []
          , Html.br [] []
          ]
        ]



-- HTTP


usersListDecoder : Decoder (List User)
usersListDecoder =
    list userDecoder


loadUsers : Model -> Cmd Msg
loadUsers model =
    Http.request
        { method = "GET"
        , url = "/api/users"
        , expect = Http.expectJson LoadedUsers usersListDecoder
        , headers = [ authHeader model.session.loginToken ]
        , body = emptyBody
        , timeout = Nothing
        , tracker = Nothing
        }
