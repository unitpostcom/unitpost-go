# unitpost-go

Official Unitpost SDK for Go.

```bash
go get github.com/unitpostcom/unitpost-go
```

```go
client := unitpost.New() // reads UNITPOST_API_KEY
data, err := client.Email.Send(ctx, map[string]any{
    "from": "Acme <team@acme.com>",
    "to":   "user@example.com",
    "html": "<p>Hi</p>",
})
```

Docs: https://www.unitpost.com/docs
