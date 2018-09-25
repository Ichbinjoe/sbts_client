# sbts_client

Based on the protocol outlined
[here](https://gist.github.com/Ichbinjoe/f1cc014f1beb6ceac8ab5a2524af88b5)

This was made for a class, not intended to be a serious protocol or for use for
pretty much anything.

## Compiling

You can compile the `sbts_client` binary on your system by running the following
command on a system with a working go installation:

```
go get
go build
```

This will put the `sbts_client` binary within your current directory.

## Running

To grab a file from a server, you may invoke sbts in the following manner:

```
./sbts_client get sbts://server/file
```

TLS may also be used if the server is using it by instead performing the
following:

```
./sbts_client get sbtss://server/file
```

Additional flags (documented in --help) may be necessary for proper TLS
operation.
