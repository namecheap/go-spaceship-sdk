# go-spaceship-sdk

Go client for the [Spaceship](https://www.spaceship.com/) domain & DNS API.

API reference: [docs.spaceship.dev](https://docs.spaceship.dev/).

## Install

```bash
go get github.com/namecheap/go-spaceship-sdk
```

## Usage

```go
import "github.com/namecheap/go-spaceship-sdk/client"

c, err := client.NewClient("https://spaceship.dev/api/v1", apiKey, apiSecret)
if err != nil {
	log.Fatal(err)
}

domains, err := c.GetDomainList(ctx)
```

Authentication uses the `X-API-Key` / `X-API-Secret` headers; credentials are
created in the Spaceship [API Manager](https://www.spaceship.com/application/api-manager/).

## License

Apache 2.0 — see [LICENSE](LICENSE).
