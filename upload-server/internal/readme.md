# Internal Packages

## OAuth Token Verification Middleware

- package location: `upload-server/internal/middleware/authverification.go`

This middleware provides OAuth 2.0 token verification for incoming requests. It currently supports JWT tokens with the plan to add support for opaque tokens. You can use it to protect either your entire router or individual routes.

### Configuration

You need to configure the middleware by setting up the following environment variables for your OAuth settings:

```shell
OAUTH_AUTH_ENABLED=true            # Enable or disable OAuth token verification
OAUTH_ISSUER_URL=https://issuer.url # URL of the token issuer
OAUTH_REQUIRED_SCOPES="scope1 scope2" # Space-separated list of required scopes
OAUTH_INTROSPECTION_URL=https://introspection.url # (for opaque tokens)
```

Then you must create an instance of the AuthHandler struct

```go
authMiddleware := middleware.AuthMiddleware{
  AuthEnabled:    appConfig.OauthConfig.AuthEnabled,
  IssuerUrl:      appConfig.OauthConfig.IssuerUrl,
  RequiredScopes: appConfig.OauthConfig.RequiredScopes,
 }
```

### Usage

#### Wrapping and Protecting the Entire Router

```go
func GetRouter(uploadUrl string, infoUrl string) http.Handler {
    router := http.NewServeMux()

    router.HandleFunc("/route-1", route1Handler)
    router.HandleFunc("/route-2", route2Handler)

    // Wrap the router with the OAuth middleware
    protectedRouter := authMiddleware.VerifyOAuthTokenMiddleware(router)

    // Return the wrapped router (as http.Handler)
    return protectedRouter
}
```

#### Wrapping and Protecting an Individual Route

```go
func GetRouter(uploadUrl string, infoUrl string) http.Handler {
    router := http.NewServeMux()

    // Wrap the particular route that needs to be protected
    router.HandleFunc("/public-route", publicRouteHandler)
    router.HandleFunc("/private-route", authMiddleware.VerifyOAuthTokenHandler(privateRouteHandler))

    return router
}
```
