# enpeeem
Serves a local npm registry with the option to proxy calls to a remote registry if packages are not available locally.

## Getting started
Start enpeeem with proxy stashing enabled to proxy requests to a remote npm registry if the request is not found locally. Which is the case when starting out.

```shell
mkdir storage
enpeeem --proxy-stash storage
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
enpeeem storage
```
