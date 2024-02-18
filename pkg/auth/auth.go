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
	GetRequestToken() string
	GetGuardModel() IGuard
	GetAccessTokenModel() IAccessToken
}

type Auth struct {
	RequestToken     string `json:"token"`
	AccessTokenModel IAccessToken
	GuardModel       IGuard
}

var _ IAuth = (*Auth)(nil)

func (a *Auth) GetRequestToken() string {
	return a.RequestToken
}

func (a *Auth) GetGuardModel() IGuard {
	return a.GuardModel
}

func (a *Auth) GetAccessTokenModel() IAccessToken {
	return a.AccessTokenModel
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
