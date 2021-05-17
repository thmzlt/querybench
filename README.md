# querybench

This is a typical Go project, so build it with `go mod vendor && go build .`. It was developed
with Go 1.16 on Linux 64-bit.

Run `querybench -h` to see the help menu. You want point it to a query CSV file
via a command line argument and to a Postgres database (set up from the task
instructions) via the `DATABASE_URL` environment variable. For example:

```
DATABASE_URL=postgres://postgres@localhost/homework ./querybench -f query_params.csv
```
