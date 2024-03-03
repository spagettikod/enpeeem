# enpeeem
Serves a local npm registry with the option to proxy calls to a remote registry if packages are not available locally.

## Install
enpeemem is a single binary and can be installed using these instructions:
<details>
<summary>macOS (Apple Silicon)</summary>

```shell
curl -OL https://github.com/spagettikod/enpeeem/releases/download/1.1.0/enpeeem1.1.0.macos-arm64.tar.gz
sudo tar -C /usr/local/bin -xvf enpeeem1.1.0.macos-arm64.tar.gz
```
</details>

<details>
<summary>Linux</summary>

```shell
curl -OL https://github.com/spagettikod/enpeeem/releases/download/1.1.0/enpeeem1.1.0.linux-amd64.tar.gz
sudo tar -C /usr/local/bin -xvf enpeeem1.1.0.linux-amd64.tar.gz
```
</details>

## Getting started
Start enpeeem in proxy mode to proxy requests to a remote npm registry. If the request resource is not found locally it's downloaded and stored.

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

## Usage
```
Local npm registry and proxy.

Packages are served from the given path. Run in proxy mode to download from
remote registry and save tarballs if they are not found locally at path.

Usage:
  enpeeem [flags] <path>

Flags:
  -addr string
        network address of local registry (default ":8080")
  -fetch-all
        download all tarbal versions at once if a tarball is not found locally
  -index string
        re-index with given package URI, example registry.npmjs.org/@types/react
  -index-all
        re-index all packages
  -progress
        show progress where applicable
  -proxystash
        run in proxy mode to proxy and download tarballs if not available locally
  -registry string
        remote npm registry to use when the flag proxystash is set (default "https://registry.npmjs.org")
  -verbose
        print debug information
  -version
        print version
```

## Indexing
enpeeem maintains package metadata files, these files are stored in each package folder as `metadata.json`.

When running enpeeem as a proxy the package metadata is automatically maintained. If new tarballs are requested and not found locally they are downloaded to local storage and the metadata file is reindexed with the new tarball.

### Manual indexing
If you remove or add tarballs manually you can trigger a manual reindexing by calling the endpoint `/api/index/<registry>/<package>`.

Example:
```
curl -X POST localhost:8080/api/index/registry.npmjs.org/typescript 
```

or for scoped packages:
```
curl -X POST 'localhost:8080/api/index/registry.npmjs.org/@types%2Freact'
```

By default the call is synchronous and will return when indexing is done. To make it asynchronous and start indexing in the background set the `async` query parameter to `true`.
```
curl -X POST 'localhost:8080/api/index/registry.npmjs.org/@types%2Freact?async=true'
```

The indexing API always responds with an empty body and one of the following HTTP status codes:
```
202 Accepted                  - Asynchronous indexing started, errors can be found in the log.
204 No Content                - Synchronous indexing completed without any errors.
429 Too Many Requests         - Package is currently being reindexed.
500 Internal Server Error     - Unexpected error occured, check the log.
```

You can also run enpeeem with the `-index-all` or `-index` flags to reindex all packages or a single package.

When package metadata is reindexed it's content is syncronized with tarballs found in storage. New tarballs are added to metadata and tarballs no longer found in storage are removed from metadata. If you want to reeindex all tarballs forcefully you need to remove the `metadata.json` file(s).

### Auto-indexing
If running in local mode, only serving local files, it reads the local metadata files. If there is no metadata file it checks for tarballs. If there are tarballs they are indexed and a metadata file is created and saved for upcoming requests.

### Indexing errors
If metadata can not be read from the tarball, for some reason, errors are logged to output. The metadata file is still created but without tarballs that could not be read.
