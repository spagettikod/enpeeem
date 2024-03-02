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
When running enpeeem as a proxy the index is automatically maintained. If enpeeem find tarballs but no metadata file it creates the metadata file and saves it to disk. This file is then used for consecutive requests. If it can not read all tarballs those that could be read are added to the metadata and returned, but the file is not saved. Check your application output to find tarballs that can not be read.

When you index a package with an existing metadata file only new files are added. Deleted tarballs are not removed from the metadata file. If you need to remove or re-index already indexed files you need to remove the metadata file and enpeeem will reindex it the next time it's requested.
