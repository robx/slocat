# slocat: proxy sockets with extra latency

`slocat` copies data between tcp and/or unix sockets, adding a configurable
delay (to both directions).

For example, to simulate a 10ms ping to your postgres server:

```
$ slocat -delay 5ms -dst localhost:5432 -src localhost:5433 &
$ psql "host=localhost port=5433"
```

or

```
$ slocat -delay 5ms -dst /tmp/.s.PGSQL.5432 -src /tmp/.s.PGSQL.5433 &
$ psql "port=5433"
```
