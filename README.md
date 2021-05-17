# querybench

To build `querybench`, make sure you have the [Nix package
manager](https://nixos.org) installed, and then run `nix-build`. This will put
the a statically-built binary under `result/bin/querybench`. You can also build
a container image with `nix-build image.nix` which can be loaded into Docker
with `docker load < result`.

Run `querybench -h` to see the help menu. You want point it to a query CSV file
via a command line argument and to a Postgres database (set up from the task
instructions) via the `DATABASE_URL` environment variable. For example:

```
DATABASE_URL=postgres://postgres@localhost/homework ./querybench -f query_params.csv
```
