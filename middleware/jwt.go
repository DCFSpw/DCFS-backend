package middleware

import (
	"dcfs/constants"
	"dcfs/responses"
	"dcfs/util/logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
)

type UserData struct {
	UserUUID uuid.UUID
}

type JWTClaim struct {
	UUID  uuid.UUID `json:"uuid"`
	Email string    `json:"email"`
	jwt.StandardClaims
}

// jwtKey - secret key used to sign JTW tokens
var jwtKey = []byte("DCFS_JWT_KEY")

// GenerateToken - generate JWT token for the user
//
// This function generated JWT token which contains UUID and e-mail of
// the requesting user. Token is then signed using secret JWT key, which
// guarantees integrity of the token on authentication.
//
// params:
//   - uuid uuid.UUID: UUID of the requesting user
//   - email string: email of the requesting user
//
// return type:
//   - signedToken string: JWT token signed using secret JWT key
//   - err error: error if signing failed, nil otherwise
func GenerateToken(uuid uuid.UUID, email string) (signedToken string, err error) {
	// Create the claims
	expirationTime := time.Now().Add(constants.JWT_TOKEN_EXPIRATION_TIME)
	claims := &JWTClaim{
		UUID:  uuid,
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	signedToken, err = token.SignedString(jwtKey)
	return
}

func validateToken(signedToken string) (claims *JWTClaim, errCode string) {
	// Parse the token
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return nil, constants.AUTH_JWT_INVALID
	}

	// Retrieve the claims
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return nil, constants.AUTH_JWT_INVALID
	}

	// Check if the token is expired
	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, constants.AUTH_JWT_EXPIRED
	}

	return claims, constants.SUCCESS
}

// Authenticate - authenticate user using JWT token
//
// This function provides functionality of JWT token authentication for
// incoming API requests. It's used by Gin engine as one of the middlewares.
// It retrieves the bearer token from the request and validates it.
// If the token is valid and not expired, it saves the user UUID (embedded in
// the token) in the context of the request. Validation whether the user with
// such UUID exists in the database or is owner of the requested resource
// is performed in the request handlers on the need basis.
//
// return type:
//   - gin.HandlerFunc: gin middleware function for authentication
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from the header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			logger.Logger.Error("middleware", "The request was unauthorized.")
			c.JSON(401, responses.NewOperationFailureResponse(constants.AUTH_JWT_MISSING, "Unauthorized"))
			c.Abort()
			return
		}

		// Check if the token is bearer token and if it is, remove the bearer part
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			logger.Logger.Error("middleware", "The request was unauthorized.")
			c.JSON(401, responses.NewOperationFailureResponse(constants.AUTH_JWT_NOT_BEARER, "Unauthorized"))
			c.Abort()
			return
		}
		tokenString = tokenString[7:]

		// Validate the token
		claims, errCode := validateToken(tokenString)
		if errCode != constants.SUCCESS {
			logger.Logger.Error("middleware", "The request was unauthorized.")
			c.JSON(401, responses.NewOperationFailureResponse(errCode, "Unauthorized"))
			c.Abort()
			return
		}

		// Set the user data in the context
		c.Set("UserData", UserData{UserUUID: claims.UUID})
	}
}
