# What is `pg-to-es`?

pg-to-es - is a project housing 2 binaries - a `pipeline` to sync CRUD operations from postgresql to elasticsearch and an `http server` exposing REST APIs for querying data thus indexed by `pipeline` in elasticsearch.<br>You can view the endpoints and their descriptions by visiting http://localhost:8080.

## How to use `pg-to-es`?

Both the binaries are self contained and ready to be run by using `Docker`, which means you need to have `Docker` up and `running`. At bare minimum it needs an environment file named `.env`, (which is by-default created when you use `make up`) with following contents.

`PostgresSQL` schema will be created (during `make up`) but database will not be seeded. This is intentional, as `elasticsearch` containers is not ready by the time `seed` runs in `postgresql`. See [Improvements](#improvements)


**Thus we have to manually seed the data by using contents from [000002_seed.up.sql](internal/db/migrations/000002_seed.up.sql)**


```.env
PG_HOST=postgres # postgresql host
PG_PORT=4532 # postgresql post
PG_USERNAME=user # postgresql username
PG_PASSWORD=secret # postgresql password
PG_DB_NAME=db # postgresql database name
PG_LISTENER_CHANNEL=core_db_event # channel to listen delta from postgresql
ES_HOST=http://elasticsearch:9200 # elasticsearch host
ES_INDEX=root # elasticsearch index
SERVER_PORT=8080 # api server port
```

###  

To run the app using `Docker` just type

```sh
  make up
```

to run tests

```sh
  make test
```

to tear down

```sh
  make down
```

<a id="improvements"></a>
### Improvements
Use shock absorber (`Message Queue`) in pipeline to retain delta during all in one boot up (`make up`).