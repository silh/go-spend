package authentication

import (
	"github.com/gofrs/uuid"
	"go-spend/authentication/jwt"
	"go-spend/expenses"
	"go-spend/util"
	"time"
)

// Performs user authentication
type AuthService interface {
	Authenticate(name util.NonEmptyString, password expenses.Password)
}

type DefaultAuthService struct {
	accessSecret  string
	refreshSecret string
}

type TokenInfo struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

func (a *DefaultAuthService) CreateToken(userid uint64) (TokenInfo, error) {
	td := TokenInfo{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	accessUUID, err := uuid.NewV4()
	if err != nil {
		return TokenInfo{}, err
	}
	td.AccessUUID = accessUUID.String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	refreshUUID, err := uuid.NewV4()
	if err != nil {
		return TokenInfo{}, err
	}
	td.RefreshUUID = refreshUUID.String()

	//Creating Access Token
	atClaims := jwt.NewClaims()
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	algorithm := jwt.HmacSha256(a.accessSecret) // TODO create just once
	td.AccessToken, err = algorithm.Encode(atClaims)
	if err != nil {
		return TokenInfo{}, err
	}
	//Creating Refresh Token
	rtClaims := jwt.NewClaims()
	rtClaims["refresh_uuid"] = td.RefreshUUID
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rtAlg := jwt.HmacSha256(a.refreshSecret)
	td.RefreshToken, err = rtAlg.Encode(rtClaims)
	if err != nil {
		return TokenInfo{}, err
	}
	return td, nil
}
