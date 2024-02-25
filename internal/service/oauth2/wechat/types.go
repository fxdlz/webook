package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"webook/internal/domain"
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type service struct {
	appID     string
	appSecret string
	client    *http.Client
}

func NewService(appID string, appSecret string) Service {
	return &service{
		appID:     appID,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

const redirectURL = ``

const authURLPattern = `https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect`
const accessTokenURLPattern = `https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code`

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	return fmt.Sprintf(authURLPattern, s.appID, redirectURL, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	url := fmt.Sprintf(accessTokenURLPattern, s.appID, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("调用微信接口失败 errcode %d, errmsg %s", res.ErrCode, res.ErrMsg)
	}
	return domain.WechatInfo{
		UnionId: res.UnionId,
		OpenId:  res.OpenId,
	}, nil
}

type Result struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}
