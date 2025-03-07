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
OAUTH_SESSION_KEY=123abc... # 32 byte or longer random string that is used to hash user session cookies
OAUTH_SESSION_DOMAIN=mydomain.com # sets the Domain setting of the user session cookie.  Useful if the UI and Upload servers live on different subdomains.
```

Then you must create an instance of the AuthHandler struct

```go
authMiddleware := NewAuthMiddleware(ctx, appConfig.OauthConfig)
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
    router.HandleFunc("/private-route", authMiddleware.VerifyOAuthTokenMiddleware(privateRouteHandler))

    return router
}
```

## User Session Cookies
This program uses the gorilla/sessions package to instantate and manage user sessions.  Sessions are stored in browser cookies.  These sessions hold access tokens as well as redirect URLs for end users, and are used to protect certain pages of the front end user interface that should only be accessed by an authenticated user.  In addition, it is used to set Authorization headers in requests to the upload server.  The following UX and security features are also implemented:

1. Cookie hashing using a secret key.  This ensures server only accepts cookies created by the server.
2. Secure and HTTPOnly enabled by default.  This prevents cookies from getting leaked due to XSS, and makes it extremely difficult for cookies to get intercepted in MitM attacks.
3. Cookie expires at the same time as their JWT.
4. Automatic user redirect.  Unauthenticated users are redirected to the login page when trying to access a protected page, and then automatically redirected to their original destination once logged in.

The following are known security risks and future improvements:

1. Cookie can still be manipulated by end user.  Poses risk if user device is compromised.  Cookie can be read from user's hard drive and leaked.
2. Cookie domain setting sets cdc.gov domain.  This allows cookies to be sent to any cdc.gov subdomain, when it should only go to the server itself.
3. Proper "login with provider" buttons as opposed to inputting a raw JWT
4. Use an external storage solution like Redis to store the JWT instead of storing it in the cookie itself.  Cookie can instead store a simple session ID.