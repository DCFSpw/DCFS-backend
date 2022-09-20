package middleware

import (
	"dcfs/constants"
	"dcfs/responses"
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

var jwtKey = []byte("DCFS_JWT_KEY")

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

func ValidateToken(signedToken string) (claims *JWTClaim, errCode string) {
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

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the JWT token from the header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Unauthorized", Code: constants.AUTH_JWT_MISSING})
			c.Abort()
			return
		}

		// Validate the token
		claims, errCode := ValidateToken(tokenString)
		if errCode != constants.SUCCESS {
			c.JSON(401, responses.InvalidCredentialsResponse{Success: false, Message: "Unauthorized", Code: errCode})
			c.Abort()
			return
		}

		// Set the user data in the context
		c.Set("UserData", UserData{UserUUID: claims.UUID})
	}
}
