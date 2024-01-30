package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

func init() {
	gob.Register(&Auth{})
}

type IAuth interface {
	GetShopID() int64
	GetMerchantID() int64 // string是为了兼容历史
	GetRequestToken() string
	GetTokenModel() IAccessToken
	IGuard
	GetGuardModel() IGuardModel
}

type Auth struct {
	ShopID          int64        `json:"shop_id"`
	MerchantID      int64        `json:"merchant_id"`
	RequestToken    string       `json:"token"`
	TokenModel      IAccessToken `json:"-" yaml:"-"` // not export to json and yaml
	GuardName       string       `json:"guard_name"`
	AuthorizationID int64        `json:"authorization_id"`
	GuardModel      IGuardModel  `json:"-" yaml:"-"` // not export to json and yaml
}

var _ IAuth = (*Auth)(nil)

func (a *Auth) GetMerchantID() int64 {
	return a.MerchantID
}

func (a *Auth) GetShopID() int64 {
	return a.ShopID
}

func (a *Auth) GetRequestToken() string {
	return a.RequestToken
}

func (a *Auth) GetTokenModel() IAccessToken {
	return a.TokenModel
}

func (a *Auth) GetGuardName() string {
	return a.GuardName
}

func (a *Auth) GetAuthorizationID() int64 {
	return a.AuthorizationID
}

func (a *Auth) GetGuardModel() IGuardModel {
	return a.GuardModel
}

func (a *Auth) MarshalBinary() ([]byte, error) {
	type _Auth Auth

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode((*_Auth)(a)); err != nil {
		return nil, err
	}

	s := base64.StdEncoding.EncodeToString(buf.Bytes())
	return []byte(s), nil
}

func (a *Auth) UnmarshalBinary(data []byte) error {
	type _Auth Auth
	var _a _Auth

	buf, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	if err = gob.NewDecoder(bytes.NewBuffer(buf)).Decode(&_a); err != nil {
		return err
	}

	*a = Auth(_a)
	return nil
}
