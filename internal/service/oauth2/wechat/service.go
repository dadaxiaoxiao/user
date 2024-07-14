package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dadaxiaoxiao/user/internal/domain"
	"net/http"
	"net/url"
)

var redirectURI = url.PathEscape("https://qinyeyiyi.cn/oauth2/wechat/callback")

type service struct {
	appId     string
	appSecret string
	client    *http.Client
}

func Newservice(appid string, appSecret string) Service {
	return &service{
		appId:     appid,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (s *service) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

func (s *service) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)

	// 构建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	defer resp.Body.Close()

	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}

	if res.ErrCode != 0 {
		// 错误返回
		return domain.WechatInfo{}, fmt.Errorf("微信返回错误响应，错误码%d,错误信息%s", res.ErrCode, res.ErrMsg)
	}

	return domain.WechatInfo{
		OpenId:  res.OpenId,
		UnionId: res.UnionId,
	}, err
}

type Result struct {
	// 错误返回
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	// 授权部分
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	// ID 部分
	OpenId  string `json:"openid"` //授权用户唯一标识
	Scope   string `json:"scope"`
	UnionId string `json:"unionid"`
}
