# enpeeem
Serves a local npm registry with the option to proxy calls to a remote registry if packages are not available locally.

## Install
enpeemem is a single binary and can be installed using these instructions:
<details>
<summary>macOS (Apple Silicon)</summary>

```shell
curl -OL https://github.com/spagettikod/enpeeem/releases/download/1.0.0/enpeeem1.0.0.macos-arm64.tar.gz
sudo tar -C /usr/local/bin -xvf enpeeem1.0.0.macos-arm64.tar.gz
```
</details>

<details>
<summary>Linux</summary>

```shell
curl -OL https://github.com/spagettikod/enpeeem/releases/download/1.0.0/enpeeem1.0.0.linux-amd64.tar.gz
sudo tar -C /usr/local/bin -xvf enpeeem1.0.0.linux-amd64.tar.gz
```
</details>

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

### Indexing
enpeeem maintans package metadata files that are returned to package managers on request. These files are stored in each package folder as `metadata.json`.

When running enpeeem as a proxy the index is automatically maintained. If new tarballs are requested and not found locally they are downloaded to local storage and the metadata file is reindexed with the new tarball.

When a metadata file is reindexed only new files are added, nothing is ever removed. Deleted tarballs are not removed from the metadata file. If you need to remove or re-index already indexed files you need to remove the metadata file and enpeeem will reindex it the next time it's requested.

#### Auto-indexing
If running in local mode, only serving local files, it reads the local metadata files. If there is not metadata file it checks for tarballs. If there are tarballs they are indexed and a metadata file is created and saved for upcoming requests.
When an error occurs during indexing of a tarball the metadata file is still created, without the erroneous tarball. The metadata file is then returned in the request but it's not saved for upcoming requests. The erroneous tarball is logged to output for you to follow up on.
