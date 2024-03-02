# enpeeem
Serves a local npm registry with the option to proxy calls to a remote registry if packages are not available locally.

## Getting started
Start enpeeem with proxy stashing enabled to proxy requests to a remote npm registry if the request is not found locally. Which is the case when starting out.

```shell
mkdir storage
enpeeem --proxy-stash ~/my_local_storage
```

Edit your `~/.npmrc` file and set the registry to your local enpeeem.
```ini
registry=http://localhost:8080
```

Call npm to start filling up your local npm stash.
```shell
npm create --force -y vite@latest myapp -- --template react-ts
```

Once you're done you can disconnect from the internet and keep serving your own npm registry by starting enpeeem in local mode.
```shell
enpeeem ~/my_local_storage
```

### Usage
```
Local npm registry and proxy.

Packages are served from the given path. If the flag proxypath is set the
request will be proxied to the remote registry and the result stored at
the given path.

Usage:
  enpeeem [flags] <path>

Flags:
  -addr string
        network address of local registry (default ":8080")
  -index string
        re-index with given package URI, example registry.npmjs.org/@types/react
  -index-all
        re-index all packages
  -progress
        show progress where applicable
  -proxystash
        proxy and download to storage if file is not available at storage path
  -registry string
        remote npm registry to use when the flag proxystash is set (default "https://registry.npmjs.org")
```

### Reindexing
When running enpeeem as a proxy the index is automatically maintained. Butr if you remove or add packages and/or tarballs in your local storage. You can re-index one or more packages using the `-index`-flags seen in the usage description.
