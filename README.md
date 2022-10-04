# token-exchange
CLI to authenticate to AWS Cognito user pool.
Opens a browser to login and returns access token

# Usage
Provide AWS Cognito's domain (i.e. target for Hosted UI), client id
and client secret (if required).
```shell
go run main.go cognito [COGNITO-DOMAIN] [CLIENT-ID] ([CLIENT_SECRET] if required) [flags]
```

After authentication in browser, access token is returned.
Subsequent calls will reuse refresh token unless expired.

# Build
```shell
sh ./build.sh
```


# Test
```shell
go test -v ./...
```

