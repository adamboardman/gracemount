module GeoCoord exposing (..)

import Types exposing (CoordBng, CoordLatLon)



-- Adapted from https://github.com/fmalina/BNGLatLon (public domain)


f0 : Float
f0 =
    0.9996012717



-- scale factor on the central meridian
-- The Airy 1830 ellipsoid semi-major and semi-minor axes used for OSGB36


a : Float
a =
    6377563.396


b : Float
b =
    6356256.909


e2 : Float
e2 =
    1 - (b * b) / (a * a)



-- The eccentricity


n : Float
n =
    (a - b) / (a + b)



-- The GSR80 semi-major and semi-minor axes used for WGS84


a_1 : Float
a_1 =
    6378137.0


b_1 : Float
b_1 =
    6356752.3141


e2_1 : Float
e2_1 =
    1 - (b_1 * b_1) / (a_1 * a_1)



-- The eccentricity
-- Northing & easting of true origin


n0 : Float
n0 =
    -100000.0


e0 : Float
e0 =
    400000.0


lat0 =
    49 * pi / 180


lon0 =
    -2 * pi / 180


sec_to_rad : Float -> Float
sec_to_rad x =
    x * pi / (180.0 * 3600.0)



--Perform Helmert transform between Airy 1830 and GRS80.
--forward converts from Airy 1830 to GRS80 or reverses the transformation.


helmert_transform : Float -> Float -> Float -> Bool -> ( Float, Float, Float )
helmert_transform x_1 y_1 z_1 forward =
    let
        i_s =
            (-20.4894 * 10) ^ -6

        -- The scale factor -1
        -- The translations along x, y, z axes respectively
        ( i_tx, i_ty, i_tz ) =
            ( 446.448, -125.157, 542.06 )

        -- The rotations along x, y, z respectively (in seconds)
        ( i_rxs, i_rys, i_rzs ) =
            ( 0.1502, 0.247, 0.8421 )

        s =
            if forward then
                i_s

            else
                -i_s

        ( tx, ty, tz ) =
            if forward then
                ( i_tx, i_ty, i_tz )

            else
                ( -i_tx, -i_ty, -i_tz )

        ( rxs, rys, rzs ) =
            if forward then
                ( i_rxs, i_rys, i_rzs )

            else
                ( -i_rxs, -i_rys, -i_rzs )

        rx =
            sec_to_rad rxs

        ry =
            sec_to_rad rys

        rz =
            sec_to_rad rzs

        x_2 =
            tx + (1 + s) * x_1 + -rz * y_1 + ry * z_1

        y_2 =
            ty + rz * x_1 + (1 + s) * y_1 + -rx * z_1

        z_2 =
            tz + -ry * x_1 + rx * y_1 + (1 + s) * z_1
    in
    ( x_2, y_2, z_2 )



-- Lat is obtained by an iterative procedure:


latEvolveBng : Float -> Float -> Float -> Float -> ( Float, Float )
latEvolveBng i_lat i_latOld z_2 p =
    let
        diff =
            abs (i_lat - i_latOld)

        latOld =
            i_lat

        nu =
            a / sqrt (1 - e2 * sin latOld ^ 2)

        o_lat =
            atan2 (z_2 + e2 * nu * sin latOld) p
    in
    if diff > 10 ^ -16 then
        latEvolveBng o_lat latOld z_2 p

    else
        ( o_lat, nu )



--Converts lat lon (WGS84) to british national grid (OSBG36)
--Accept latitude and longitude as used in GPS.
--Return OSGB grid coordinates: eastings and northings.


latLonToBng : CoordLatLon -> CoordBng
latLonToBng coord =
    let
        lat_1 =
            coord.lat * pi / 180.0

        lon_1 =
            coord.lon * pi / 180.0

        nu_1 =
            a_1 / sqrt (1.0 - (e2_1 * sin lat_1 ^ 2))

        --First convert to cartesian from spherical polar coordinates
        h =
            0

        -- Third spherical coord.
        x_1 =
            (nu_1 + h) * cos lat_1 * cos lon_1

        y_1 =
            (nu_1 + h) * cos lat_1 * sin lon_1

        z_1 =
            ((1.0 - e2_1) * nu_1 + h) * sin lat_1

        ( x_2, y_2, z_2 ) =
            helmert_transform x_1 y_1 z_1 False

        -- Back to spherical polar coordinates from cartesian
        -- Need some of the characteristics of the new ellipsoid
        p =
            sqrt (x_2 ^ 2 + y_2 ^ 2)

        -- Lat is obtained by an iterative procedure:
        i_lat =
            atan2 z_2 (p * (1.0 - e2))

        -- Initial value
        i_latOld =
            2 * pi

        ( lat, nu ) =
            latEvolveBng i_lat i_latOld z_2 p

        -- Lon and height are then pretty easy
        lon =
            atan2 y_2 x_2

        --H = p/cos(lat) - nu appears to never be used...?
        -- meridional radius of curvature
        rho =
            a * f0 * (1 - e2) * (1 - e2 * sin lat ^ 2) ^ -1.5

        eta2 =
            nu * f0 / rho - 1

        m1 =
            (1 + n + (5 / 4) * n ^ 2 + (5 / 4) * n ^ 3) * (lat - lat0)

        m2 =
            (3 * n + 3 * n ^ 2 + (21 / 8) * n ^ 3) * sin (lat - lat0) * cos (lat + lat0)

        m3 =
            ((15 / 8) * n ^ 2 + (15 / 8) * n ^ 3) * sin (2 * (lat - lat0)) * cos (2 * (lat + lat0))

        m4 =
            (35 / 24) * n ^ 3 * sin (3 * (lat - lat0)) * cos (3 * (lat + lat0))

        -- meridional arc
        m =
            b * f0 * (m1 - m2 + m3 - m4)

        i =
            m + n0

        -- noqa E741, use roman numerals
        ii =
            nu * f0 * sin lat * cos lat / 2

        iii =
            nu * f0 * sin lat * cos lat ^ 3 * (5 - tan lat ^ 2 + 9 * eta2) / 24

        iiia =
            nu * f0 * sin lat * cos lat ^ 5 * (61 - 58 * tan lat ^ 2 + tan lat ^ 4) / 720

        iv =
            nu * f0 * cos lat

        v =
            nu * f0 * cos lat ^ 3 * (nu / rho - tan lat ^ 2) / 6

        vi =
            nu * f0 * cos lat ^ 5 * (5 - 18 * tan lat ^ 2 + tan lat ^ 4 + 14 * eta2 - 58 * eta2 * tan lat ^ 2) / 120

        -- E, N are the British national grid coordinates - eastings and northings
        northing =
            i + ii * (lon - lon0) ^ 2 + iii * (lon - lon0) ^ 4 + iiia * (lon - lon0) ^ 6

        easting =
            e0 + iv * (lon - lon0) + v * (lon - lon0) ^ 3 + vi * (lon - lon0) ^ 5
    in
    { easting = easting
    , northing = northing
    }


latEvolveMeridional : Float -> Float -> ( Float, Float )
latEvolveMeridional i_lat i_m =
    let
        lat =
            (n - n0 - i_m) / (a * f0) + i_lat

        m1 =
            (1 + n + (5.0 / 4) * n ^ 2 + (5.0 / 4) * n ^ 3) * (lat - lat0)

        m2 =
            (3 * n + 3 * n ^ 2 + (21.0 / 8) * n ^ 3) * sin (lat - lat0) * cos (lat + lat0)

        m3 =
            ((15.0 / 8) * n ^ 2 + (15.0 / 8) * n ^ 3) * sin (2 * (lat - lat0)) * cos (2 * (lat + lat0))

        m4 =
            (35.0 / 24) * n ^ 3 * sin (3 * (lat - lat0)) * cos (3 * (lat + lat0))

        -- meridional arc
        m =
            b * f0 * (m1 - m2 + m3 - m4)

        diff =
            n - n0 - m
    in
    if diff >= 0.00001 then
        latEvolveMeridional lat m

    else
        ( lat, m )


latEvolveBng1 : Float -> Float -> Float -> Float -> ( Float, Float )
latEvolveBng1 i_lat i_latOld z_2 p =
    let
        diff =
            abs (i_lat - i_latOld)

        latOld =
            i_lat

        nu_2 =
            a_1 / sqrt (1 - e2_1 * sin latOld ^ 2)

        o_lat =
            atan2 (z_2 + e2_1 * nu_2 * sin latOld) p
    in
    if diff > 10 ^ -16 then
        latEvolveBng o_lat latOld z_2 p

    else
        ( o_lat, nu_2 )



--Converts british national grid to lat lon
--Accepts The Ordnance Survey National Grid eastings and northings.
--Return latitude and longitude coordinates.


bngToLatLon : CoordBng -> CoordLatLon
bngToLatLon coord =
    let
        ( lat, m ) =
            latEvolveMeridional lat0 0

        -- transverse radius of curvature
        nu =
            a * f0 / sqrt (1 - e2 * sin lat ^ 2)

        -- meridional radius of curvature
        rho =
            a * f0 * (1 - e2) * (1 - e2 * sin lat ^ 2) ^ -1.5

        eta2 =
            nu / rho - 1

        sec_lat =
            1.0 / cos lat

        vii =
            tan lat / (2 * rho * nu)

        viii =
            tan lat / (24 * rho * nu ^ 3) * (5 + 3 * tan lat ^ 2 + eta2 - 9 * tan lat ^ 2 * eta2)

        ix =
            tan lat / (720 * rho * nu ^ 5) * (61 + 90 * tan lat ^ 2 + 45 * tan lat ^ 4)

        x =
            sec_lat / nu

        xi =
            sec_lat / (6 * nu ^ 3) * (nu / rho + 2 * tan lat ^ 2)

        xii =
            sec_lat / (120 * nu ^ 5) * (5 + 28 * tan lat ^ 2 + 24 * tan lat ^ 4)

        xiia =
            sec_lat / (5040 * nu ^ 7) * (61 + 662 * tan lat ^ 2 + 1320 * tan lat ^ 4 + 720 * tan lat ^ 6)

        dE =
            e - e0

        -- These are on the wrong ellipsoid currently: Airy 1830 (denoted by _1)
        lat_1 =
            lat - vii * dE ^ 2 + viii * dE ^ 4.0 - ix * dE ^ 6.0

        lon_1 =
            lon0 + x * dE - xi * dE ^ 3 + xii * dE ^ 5.0 - xiia * dE ^ 7.0

        -- Want to convert to the GRS80 ellipsoid.
        --First convert to cartesian from spherical polar coordinates
        h =
            0

        -- Third spherical coord.
        x_1 =
            (nu / f0 + h) * cos lat_1 * cos lon_1

        y_1 =
            (nu / f0 + h) * cos lat_1 * sin lon_1

        z_1 =
            ((1 - e2) * nu / f0 + h) * sin lat_1

        ( x_2, y_2, z_2 ) =
            helmert_transform x_1 y_1 z_1 True

        -- Back to spherical polar coordinates from cartesian
        -- Need some of the characteristics of the new ellipsoid
        p =
            sqrt (x_2 ^ 2 + y_2 ^ 2)

        -- Lat is obtained by an iterative procedure:
        lat_2 =
            atan2 z_2 (p * (1 - e2_1))

        -- Initial value
        latOld =
            2 * pi

        ( lat_3, nu_2 ) =
            latEvolveBng1 lat_2 latOld z_2 p

        -- Lon and height are then pretty easy
        lon =
            atan2 y_2 x_2

        --h = p/cos(lat) - nu_2
        -- Uncomment this line if you want to print the results
        -- print([(lat-lat_1)*180/pi, (lon - lon_1)*180/pi])
        -- Convert to degrees
        lat_o =
            lat * 180.0 / pi

        lon_o =
            lon * 180.0 / pi
    in
    { lat = lat_o
    , lon = lon_o
    }



--# tests
--assert latlon_to_bng(51.4778, -0.0014) == (538890.1053, 177320.4965)
--assert latlon_to_bng(53.50713, -2.71766) == (352500.1952, 401400.0148)
--assert bng_to_latlon(538890, 177320) == (51.477795, -0.001402)
--assert bng_to_latlon(352500.2, 401400) == (53.50713, -2.71766)
