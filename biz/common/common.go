package common

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"code.byted.org/qiumingliang.123/personal_bot/biz/model"
	"code.byted.org/qiumingliang.123/personal_bot/conf"
)

// APIPath
var (
	TenantAccessTokenURL = "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
)

// GetTenantAccessToken get tenant access token for app
// Refer to: https://open.feishu.cn/document/ukTMukTMukTM/ukDNz4SO0MjL5QzM/auth-v3/auth/tenant_access_token_internal
func GetTenantAccessToken(ctx context.Context) (string, error) {
	cli := &http.Client{}
	reqBody := model.TenantAccessTokenRequest{
		APPID:     conf.Conf.AppID,
		APPSecret: conf.Conf.AppSecret,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", TenantAccessTokenURL, strings.NewReader(string(reqBytes)))
	if err != nil {
		return "", err
	}
	resp, err := cli.Do(req)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get token")
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenResp := &model.TenantAccessTokenResponse{}
	err = json.Unmarshal(body, tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.TenantAccessToken, nil
}

func GetStrText(ss string) string {
	s := []rune(ss)
	str := make([]rune, 0)
	i := 0
	j := len(s) - 1
	for ; i <= j; i++ {
		if s[i] != ' ' && s[i] != '\n' {
			break
		}
	}
	for ; j >= i; j-- {
		if s[j] != ' ' && s[j] != '\n' {
			break
		}
	}
	for k := i; k <= j; k++ {
		str = append(str, s[k])
	}
	return string(str)
}

func IsEqualString(s1, s2 string) bool {
	return GetStrText(s1) == GetStrText(s2)
}
