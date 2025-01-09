# airvantage-api-go

Go client for AirVantage device management REST API.

## Release manually a new version

As Go uses a [specific version format](https://go.dev/doc/modules/version-numbers) we cannot use the usual `YY.MM.<counter>` numbering scheme. We can use `v1.YYMM..<counter>` instead.
You simply need to create a tag and push it

```sh
git tag v1.2501.1
git push origin tag v1.2501.1
```

Then you can fetch the version in  other Go project

```sh
go get github.com/AirVantage/airvantage-api-go@v1.2501.1
```
