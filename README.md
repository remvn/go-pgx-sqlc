Source code of [this article](https://remvn.dev/posts/raw-dogging-postgresql-with-pgx-and-sqlc-in-go/)

### How to run:
1. Install `docker` on your system
2. Run this command:
```bash
go test ./... 
```

With the implementation of [testcontainer for
go](https://golang.testcontainers.org/), it will create a postgres container on
the fly and run integration test on that database!

