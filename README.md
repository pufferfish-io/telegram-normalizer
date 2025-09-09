# telegram-normalizer

```
export $(cat .env | xargs) && go run ./cmd/telegram-normalizer
```

```
go mod tidy
```

```
go build -v -x ./cmd/telegram-normalizer && rm -f telegram-normalizer
```

```
docker buildx build --no-cache --progress=plain .
```

```
set -a && source .env && set +a && go run ./cmd/telegram-normalizer
```

```
git tag v0.1.1
git push origin v0.1.1
```

```
git tag -l
git tag -d vX.Y.Z
git push --delete origin vX.Y.Z
git ls-remote --tags origin | grep 'refs/tags/vX.Y.Z$'
```
