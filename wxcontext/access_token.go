package wxcontext

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/radishqiu/gowechat/util"
)

const (
	//AccessTokenURL 获取access_token的接口
	AccessTokenURL = "https://api.weixin.qq.com/cgi-bin/token"
)

//ResAccessToken struct
type ResAccessToken struct {
	util.CommonError

	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

//SetAccessTokenLock 设置读写锁（一个appID一个读写锁）
func (ctx *Context) SetAccessTokenLock(l *sync.RWMutex) {
	ctx.accessTokenLock = l
}

//GetAccessToken 获取access_token
func (ctx *Context) GetAccessToken() (accessToken string, err error) {
	ctx.accessTokenLock.Lock()
	defer ctx.accessTokenLock.Unlock()

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	val := ctx.Cache.Get(accessTokenCacheKey)
	if val != nil {
		accessToken = val.(string)
		return
	}

	//从微信服务器获取
	var resAccessToken ResAccessToken
	resAccessToken, err = ctx.GetAccessTokenFromServer()
	if err != nil {
		return
	}

	accessToken = resAccessToken.AccessToken
	return
}

//CleanAccessTokenCache clean cache
func (ctx *Context) CleanAccessTokenCache() {
	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	ctx.Cache.Delete(accessTokenCacheKey)
}

//GetAccessTokenFromServer 强制从微信服务器获取token
func (ctx *Context) GetAccessTokenFromServer() (resAccessToken ResAccessToken, err error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", AccessTokenURL, ctx.AppID, ctx.AppSecret)
	var body []byte
	body, err = util.HTTPGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrMsg != "" {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
		return
	}

	accessTokenCacheKey := fmt.Sprintf("access_token_%s", ctx.AppID)
	expires := resAccessToken.ExpiresIn - 1500
	err = ctx.Cache.Put(accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(expires)*time.Second)
	return
}
