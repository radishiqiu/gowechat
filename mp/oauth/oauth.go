package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/radishqiu/gowechat/util"
	"github.com/radishqiu/gowechat/wxcontext"
)

const (
	redirectOauthURL      = "https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect"
	accessTokenURL        = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	refreshAccessTokenURL = "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s"
	userInfoURL           = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=zh_CN"
	checkAccessTokenURL   = "https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s"
)

//Oauth 保存用户授权信息
type Oauth struct {
	*wxcontext.Context
}

//NewOauth 实例化授权信息
func NewOauth(context *wxcontext.Context) *Oauth {
	auth := new(Oauth)
	auth.Context = context
	return auth
}

//GetRedirectURL 获取跳转的url地址
func (oauth *Oauth) GetRedirectURL(redirectURI, scope, state string) string {
	//url encode
	urlStr := url.QueryEscape(redirectURI)
	return fmt.Sprintf(redirectOauthURL, oauth.AppID, urlStr, scope, state)
}

//Redirect 跳转到网页授权
func (oauth *Oauth) Redirect(redirectURI, scope, state string) error {
	location := oauth.GetRedirectURL(redirectURI, scope, state)
	//location 为完整地址，所以不需要request
	http.Redirect(oauth.Writer, oauth.Request, location, 302)
	return nil
}

// ResAccessToken 获取用户授权access_token的返回结果
type ResAccessToken struct {
	util.CommonError

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
}

// GetUserAccessToken 通过网页授权的code 换取access_token(区别于context中的access_token)
func (oauth *Oauth) GetUserAccessToken(code string) (result ResAccessToken, err error) {
	urlStr := fmt.Sprintf(accessTokenURL, oauth.AppID, oauth.AppSecret, code)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserAccessToken error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

//RefreshAccessToken 刷新access_token
func (oauth *Oauth) RefreshAccessToken(refreshToken string) (result ResAccessToken, err error) {
	urlStr := fmt.Sprintf(refreshAccessTokenURL, oauth.AppID, refreshToken)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserAccessToken error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}

//CheckAccessToken 检验access_token是否有效
func (oauth *Oauth) CheckAccessToken(accessToken, openID string) (b bool, err error) {
	urlStr := fmt.Sprintf(checkAccessTokenURL, accessToken, openID)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	var result util.CommonError
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		b = false
		return
	}
	b = true
	return
}

//UserInfo 用户授权获取到用户信息
type UserInfo struct {
	util.CommonError

	OpenID     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	HeadImgURL string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

//GetUserInfo 如果scope为 snsapi_userinfo 则可以通过此方法获取到用户基本信息
func (oauth *Oauth) GetUserInfo(accessToken, openID string) (result UserInfo, err error) {
	urlStr := fmt.Sprintf(userInfoURL, accessToken, openID)
	var response []byte
	response, err = util.HTTPGet(urlStr)
	if err != nil {
		return
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return
	}
	if result.ErrCode != 0 {
		err = fmt.Errorf("GetUserInfo error : errcode=%v , errmsg=%v", result.ErrCode, result.ErrMsg)
		return
	}
	return
}
