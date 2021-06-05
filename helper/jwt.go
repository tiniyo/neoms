package helper

import (
	"github.com/dgrijalva/jwt-go"
	"os"
	"time"
)

type JwtTokenInfo struct {
	Ip string
	Username string
	Password string
}

type JwtTokenInfos []JwtTokenInfo

func CreateToken(jwtInfos JwtTokenInfos) (string, error) {
	var err error
	//Creating Access Token
	os.Setenv("ACCESS_SECRET", "RZ+w1Vr/dk4nZHvd/B7av/pOGiNzYlPZ") //this should be in an env file
	atClaims := jwt.MapClaims{}

	for _, jwtInfo := range jwtInfos {
		atClaims[jwtInfo.Ip] = jwtInfo.Username
		atClaims[jwtInfo.Username] = jwtInfo.Password
	}
	atClaims["authorized"] = true
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}
