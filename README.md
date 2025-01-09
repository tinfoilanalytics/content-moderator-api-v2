# content-moderator-api

This is a simple API that uses the Replicate API to analyze content.

To run the API, you need to set the `REPLICATE_API_TOKEN` environment variable.

```
export REPLICATE_API_TOKEN=r8_your_token_here
go run main.go
```

To run the API on Fly.io, you need to set the `REPLICATE_API_TOKEN` secret.

Currently, the API is deployed on Fly.io and can be accessed at https://content-moderator-api.fly.dev/api/analyze.