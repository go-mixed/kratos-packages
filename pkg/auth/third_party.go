package auth

type IThirdParty interface {
	GetID() int64
	GetAppKey() string
	GetAppSecret() string
}

type thirdParty struct {
	ID        int64  `json:"id"`
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

var _ IThirdParty = (*thirdParty)(nil)

func NewThirdParty(id int64, appKey, appSecret string) IThirdParty {
	return &thirdParty{
		ID:        id,
		AppKey:    appKey,
		AppSecret: appSecret,
	}
}

func (t *thirdParty) GetID() int64 {
	return t.ID
}

func (t *thirdParty) GetAppKey() string {
	return t.AppKey
}

func (t *thirdParty) GetAppSecret() string {
	return t.AppSecret
}
