package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials/request"
	"github.com/aliyun/credentials-go/credentials/utils"
)

// URLCredential is a kind of credential
type URLCredential struct {
	URL string
	*credentialUpdater
	*sessionCredential
	runtime *utils.Runtime
}

type URLResponse struct {
	AccessKeyId     string `json:"AccessKeyId" xml:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret" xml:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken" xml:"SecurityToken"`
	Expiration      string `json:"Expiration" xml:"Expiration"`
}

func newURLCredential(URL string) *URLCredential {
	credentialUpdater := new(credentialUpdater)
	if URL == "" {
		URL = os.Getenv("ALIBABA_CLOUD_CREDENTIALS_URI")
	}
	return &URLCredential{
		URL:               URL,
		credentialUpdater: credentialUpdater,
	}
}

// GetAccessKeyId reutrns  URLCredential's AccessKeyId
// if AccessKeyId is not exist or out of date, the function will update it.
func (e *URLCredential) GetAccessKeyId() (*string, error) {
	if e.sessionCredential == nil || e.needUpdateCredential() {
		err := e.updateCredential()
		if err != nil {
			if e.credentialExpiration > (int(time.Now().Unix()) - int(e.lastUpdateTimestamp)) {
				return &e.sessionCredential.AccessKeyId, nil
			}
			return tea.String(""), err
		}
	}
	return tea.String(e.sessionCredential.AccessKeyId), nil
}

// GetAccessSecret reutrns  URLCredential's AccessKeySecret
// if AccessKeySecret is not exist or out of date, the function will update it.
func (e *URLCredential) GetAccessKeySecret() (*string, error) {
	if e.sessionCredential == nil || e.needUpdateCredential() {
		err := e.updateCredential()
		if err != nil {
			if e.credentialExpiration > (int(time.Now().Unix()) - int(e.lastUpdateTimestamp)) {
				return &e.sessionCredential.AccessKeySecret, nil
			}
			return tea.String(""), err
		}
	}
	return tea.String(e.sessionCredential.AccessKeySecret), nil
}

// GetSecurityToken reutrns  URLCredential's SecurityToken
// if SecurityToken is not exist or out of date, the function will update it.
func (e *URLCredential) GetSecurityToken() (*string, error) {
	if e.sessionCredential == nil || e.needUpdateCredential() {
		err := e.updateCredential()
		if err != nil {
			if e.credentialExpiration > (int(time.Now().Unix()) - int(e.lastUpdateTimestamp)) {
				return &e.sessionCredential.SecurityToken, nil
			}
			return tea.String(""), err
		}
	}
	return tea.String(e.sessionCredential.SecurityToken), nil
}

// GetBearerToken is useless for URLCredential
func (e *URLCredential) GetBearerToken() *string {
	return tea.String("")
}

// GetType reutrns  URLCredential's type
func (e *URLCredential) GetType() *string {
	return tea.String("credential_uri")
}

func (e *URLCredential) updateCredential() (err error) {
	if e.runtime == nil {
		e.runtime = new(utils.Runtime)
	}
	request := request.NewCommonRequest()
	request.URL = e.URL
	request.Method = "GET"
	content, err := doAction(request, e.runtime)
	if err != nil {
		return fmt.Errorf("refresh Ecs sts token err: %s", err.Error())
	}
	var resp *URLResponse
	err = json.Unmarshal(content, &resp)
	if err != nil {
		return fmt.Errorf("refresh Ecs sts token err: Json Unmarshal fail: %s", err.Error())
	}
	if resp.AccessKeyId == "" || resp.AccessKeySecret == "" || resp.SecurityToken == "" || resp.Expiration == "" {
		return fmt.Errorf("refresh Ecs sts token err: AccessKeyId: %s, AccessKeySecret: %s, SecurityToken: %s, Expiration: %s", resp.AccessKeyId, resp.AccessKeySecret, resp.SecurityToken, resp.Expiration)
	}

	expirationTime, err := time.Parse("2006-01-02T15:04:05Z", resp.Expiration)
	e.lastUpdateTimestamp = time.Now().Unix()
	e.credentialExpiration = int(expirationTime.Unix() - time.Now().Unix())
	e.sessionCredential = &sessionCredential{
		AccessKeyId:     resp.AccessKeyId,
		AccessKeySecret: resp.AccessKeySecret,
		SecurityToken:   resp.SecurityToken,
	}

	return
}
