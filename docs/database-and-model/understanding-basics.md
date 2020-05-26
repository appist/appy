---
description: Covers how to work with the database layer.
---

# Understanding Basics

With `appy` framework, you can decide if you wanna use 1 of the below to interact with the database:

* `app.DB(name string) record.DBer` which returns a DB handle with the specified name and is using [sqlx](https://github.com/jmoiron/sqlx) as its query interface. This is recommended if you're already familiar with `database/sql` package and fine with writing boilerplate code. 
* `app.Model(dest interface{}, opts ...record.ModelOption) record.Modeler` which returns a ORM model and is using `record.DBer` as its query interface. This is recommended if you're just starting off and would like to avoid writing boilerplate code.

{% hint style="info" %}
Both `record.DBer` and `record.Modeler` are interfaces which allows us to easily provide mocks for unit testing. For more details, please refer to [Writing Unit Tests](https://app.gitbook.com/@appist/s/appy/~/drafts/-M89Lng7Tw37CmiJzM9P/database-and-model/writing-unit-tests).
{% endhint %}

### What is record.DBer?

It is a generic interface extended with [sqlx](https://github.com/jmoiron/sqlx) extension around SQL \(or SQL-like\) databases which must be used in conjunction with the database drivers:

* mysql - [https://github.com/go-sql-driver/mysql/](https://github.com/go-sql-driver/mysql/)
* postgresql - [https://github.com/lib/pq](https://github.com/lib/pq)

### What is record.Modeler, i.e. Object Relational Mapping \(ORM\)?

It is a generic interface that connects the rich objects of an application to tables in a relational database management system. Using ORM, the properties and relationships of the objects in an application can be easily stored and retrieved from a database without writing SQL statements directly and with less overall database access code.

### Benchmarks \(database/sql vs DB vs ORM\)

```bash
go test -run=NONE -bench . -benchmem -benchtime 10s -failfast ./record
goos: darwin
goarch: amd64
pkg: github.com/appist/appy/record
BenchmarkInsertRaw-4                1239          10103533 ns/op              88 B/op          5 allocs/op
BenchmarkInsertDB-4                  898          11351591 ns/op            1548 B/op         19 allocs/op
BenchmarkInsertORM-4                 826          13826999 ns/op           15338 B/op        283 allocs/op
BenchmarkInsertMultiRaw-4            529          21830643 ns/op          107896 B/op        415 allocs/op
BenchmarkInsertMultiDB-4             481          20931749 ns/op          166302 B/op        441 allocs/op
BenchmarkInsertMultiORM-4            471          23261618 ns/op          791677 B/op       3872 allocs/op
BenchmarkUpdateRaw-4                 903          13807008 ns/op            1064 B/op         21 allocs/op
BenchmarkUpdateDB-4                 1008          13577352 ns/op            3677 B/op         52 allocs/op
BenchmarkUpdateORM-4                 788          13923442 ns/op            8920 B/op        233 allocs/op
BenchmarkReadRaw-4                  2162           4723198 ns/op            1810 B/op         47 allocs/op
BenchmarkReadDB-4                   2263           5300805 ns/op            3257 B/op         69 allocs/op
BenchmarkReadORM-4                  2259           5184327 ns/op            6911 B/op        230 allocs/op
BenchmarkReadSliceRaw-4             2210           5871991 ns/op           23088 B/op       1331 allocs/op
BenchmarkReadSliceDB-4              2197           5752959 ns/op           25070 B/op       1353 allocs/op
BenchmarkReadSliceORM-4             1864           6249231 ns/op          246630 B/op       1526 allocs/op
PASS
ok      github.com/appist/appy/record   344.692s
```

