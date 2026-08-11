## Install some dependencies

```
sudo apt-get install golang postgresql postgis postgresql-postgis libvips-dev elm-compiler uglifyjs
```

## Database

Need to create postgres users and database:
```
$ sudo -u postgres psql
$ sudo -u postgres createuser gracetest
$ sudo -u postgres createdb -O gracetest gracetest
$ sudo -u postgres psql
psql=# alter user gracetest with encrypted password '<password>';
$ sudo -u postgres psql gracetest
gracetest=# CREATE EXTENSION postgis;
```

Server (similar - possibly with port incase you have multiple versions of postgres from upgrades):
```
$ sudo -u postgres psql -p 5434 grace
```

Running the go tests or main.go the first time will create a file postgres_args.txt you should edit this file with your postgres database details:
```
host=localhost port=5432 sslmode=disable user=gracetest dbname=gracetest password=[...]
```

## Go dependencies

You'll need to get lots of go dependencies using something similar to:

go get golang.org/x/sys/cpu

## Testing

Should always check that the tests are passing before committing (do not run on a live server):
```
go test ./binary
go test ./store
go test ./server
go test ./tag_updater
```

## Debugging
To run locally you need to:
```
elm make client/Main.elm --output public/elm.js --debug
go run main.go -debugging=true
```

## Compile for deployment
To compile on the server:
```
elm make client/Main.elm --optimize --output=public/elm.js
uglifyjs public/elm.js --compress 'pure_funcs="F2,F3,F4,F5,F6,F7,F8,F9,A2,A3,A4,A5,A6,A7,A8,A9",pure_getters,keep_fargs=false,unsafe_comps,unsafe' | uglifyjs --mangle --output public/elm.min.js
go build main.go
./main
```

### Add as a systemd service
/lib/systemd/system/gracemount.service
```
[Unit]
Description=gracemount

[Service]
Type=simple
Restart=always
RestartSec=5s
User=user
Group=user
WorkingDirectory=/home/user/go/src/github.com/githubuser/gracemount/
ExecStart=/home/user/go/src/github.com/githubuser/gracemount/main

[Install]
WantedBy=multi-user.target
```

## Live server config - to run on port 3040
Expected to be running via a proxy on port 80/443

## Import a trees database
If you have a spreadsheet you've previously populated you can use it to see your tree and plant catalogue with something like (copied to /tmp to avoid permissions problems):
```
$ cp importable_trees.csv /tmp
$ sudo -u postgres psql gracetest
gracetest=# \copy items(silver_number,name,description,latitude_i,longitude_i,altitude,status,age_group,height,latin_name,diameter_at,render_hints,fruit_type,cropping_season,fruit_storage,pollinating_group,rootstock) FROM '/tmp/importable_trees.csv' delimiter ',' csv header;
gracetest=# update items set view_permissions=0 where view_permissions is null;
gracetest=# update items set location=ST_SetSRID(ST_POINT(items.longitude_i/10000000.0,items.latitude_i/10000000.0), 4326) where location is null;
gracetest=# update items set item_type=2 where name like '%Apple%';
```
