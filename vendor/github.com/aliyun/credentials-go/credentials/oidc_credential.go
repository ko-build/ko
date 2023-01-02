package credentials

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials/request"
	"github.com/aliyun/credentials-go/credentials/utils"
)

const defaultOIDCDurationSeconds = 3600

// OIDCCredential is a kind of credentials
type OIDCCredential struct {
	*credentialUpdater
	AccessKeyId           string
	AccessKeySecret       string
	RoleArn               string
	OIDCProviderArn       string
	OIDCTokenFilePath     string
	Policy                string
	RoleSessionName       string
	RoleSessionExpiration int
	sessionCredential     *sessionCredential
	runtime               *utils.Runtime
}

type OIDCResponse struct {
	Credentials *credentialsInResponse `json:"Credentials" xml:"Credentials"`
}

type OIDCcredentialsInResponse struct {
	AccessKeyId     string `json:"AccessKeyId" xml:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret" xml:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken" xml:"SecurityToken"`
	Expiration      string `json:"Expiration" xml:"Expiration"`
}

func newOIDCRoleArnCredential(accessKeyId, accessKeySecret, roleArn, OIDCProviderArn, OIDCTokenFilePath, RoleSessionName, policy string, RoleSessionExpiration int, runtime *utils.Runtime) *OIDCCredential {
	return &OIDCCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleArn:               roleArn,
		OIDCProviderArn:       OIDCProviderArn,
		OIDCTokenFilePath:     OIDCTokenFilePath,
		RoleSessionName:       RoleSessionName,
		Policy:                policy,
		RoleSessionExpiration: RoleSessionExpiration,
		credentialUpdater:     new(credentialUpdater),
		runtime:               runtime,
	}
}

// GetAccessKeyId reutrns OIDCCredential's AccessKeyId
// if AccessKeyId is not exist or out of date, the function will update it.
func (r *OIDCCredential) GetAccessKeyId() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.AccessKeyId), nil
}

// GetAccessSecret reutrns OIDCCredential's AccessKeySecret
// if AccessKeySecret is not exist or out of date, the function will update it.
func (r *OIDCCredential) GetAccessKeySecret() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.AccessKeySecret), nil
}

// GetSecurityToken reutrns OIDCCredential's SecurityToken
// if SecurityToken is not exist or out of date, the function will update it.
func (r *OIDCCredential) GetSecurityToken() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.SecurityToken), nil
}

// GetBearerToken is useless OIDCCredential
func (r *OIDCCredential) GetBearerToken() *string {
	return tea.String("")
}

// GetType reutrns OIDCCredential's type
func (r *OIDCCredential) GetType() *string {
	return tea.String("oidc_role_arn")
}

func (r *OIDCCredential) GetOIDCToken(OIDCTokenFilePath string) *string {
	tokenPath := OIDCTokenFilePath
	_, err := os.Stat(tokenPath)
	if os.IsNotExist(err) {
		tokenPath = os.Getenv("ALIBABA_CLOUD_OIDC_TOKEN_FILE")
		if tokenPath == "" {
			return nil
		}
	}
	byt, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil
	}
	return tea.String(string(byt))
}

func (r *OIDCCredential) updateCredential() (err error) {
	if r.runtime == nil {
		r.runtime = new(utils.Runtime)
	}
	request := request.NewCommonRequest()
	request.Domain = "sts.aliyuncs.com"
	request.Scheme = "HTTPS"
	request.Method = "POST"
	request.QueryParams["Timestamp"] = utils.GetTimeInFormatISO8601()
	request.QueryParams["Action"] = "AssumeRoleWithOIDC"
	request.QueryParams["Format"] = "JSON"
	request.BodyParams["RoleArn"] = r.RoleArn
	request.BodyParams["OIDCProviderArn"] = r.OIDCProviderArn
	token := r.GetOIDCToken(r.OIDCTokenFilePath)
	request.BodyParams["OIDCToken"] = tea.StringValue(token)
	if r.Policy != "" {
		request.QueryParams["Policy"] = r.Policy
	}
	request.QueryParams["RoleSessionName"] = r.RoleSessionName
	request.QueryParams["Version"] = "2015-04-01"
	request.QueryParams["SignatureNonce"] = utils.GetUUID()
	if r.AccessKeyId != "" && r.AccessKeySecret != "" {
		signature := utils.ShaHmac1(request.BuildStringToSign(), r.AccessKeySecret+"&")
		request.QueryParams["Signature"] = signature
		request.QueryParams["AccessKeyId"] = r.AccessKeyId
		request.QueryParams["AccessKeySecret"] = r.AccessKeySecret
	}
	request.Headers["Host"] = request.Domain
	request.Headers["Accept-Encoding"] = "identity"
	request.Headers["content-type"] = "application/x-www-form-urlencoded"
	request.URL = request.BuildURL()
	content, err := doAction(request, r.runtime)
	if err != nil {
		return fmt.Errorf("refresh RoleArn sts token err: %s", err.Error())
	}
	var resp *OIDCResponse
	err = json.Unmarshal(content, &resp)
	if err != nil {
		return fmt.Errorf("refresh RoleArn sts token err: Json.Unmarshal fail: %s", err.Error())
	}
	if resp == nil || resp.Credentials == nil {
		return fmt.Errorf("refresh RoleArn sts token err: Credentials is empty")
	}
	respCredentials := resp.Credentials
	if respCredentials.AccessKeyId == "" || respCredentials.AccessKeySecret == "" || respCredentials.SecurityToken == "" || respCredentials.Expiration == "" {
		return fmt.Errorf("refresh RoleArn sts token err: AccessKeyId: %s, AccessKeySecret: %s, SecurityToken: %s, Expiration: %s", respCredentials.AccessKeyId, respCredentials.AccessKeySecret, respCredentials.SecurityToken, respCredentials.Expiration)
	}

	expirationTime, err := time.Parse("2006-01-02T15:04:05Z", respCredentials.Expiration)
	r.lastUpdateTimestamp = time.Now().Unix()
	r.credentialExpiration = int(expirationTime.Unix() - time.Now().Unix())
	r.sessionCredential = &sessionCredential{
		AccessKeyId:     respCredentials.AccessKeyId,
		AccessKeySecret: respCredentials.AccessKeySecret,
		SecurityToken:   respCredentials.SecurityToken,
	}

	return
}
