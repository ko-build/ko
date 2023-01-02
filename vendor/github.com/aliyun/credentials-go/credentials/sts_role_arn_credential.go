package credentials

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials/request"
	"github.com/aliyun/credentials-go/credentials/utils"
)

const defaultDurationSeconds = 3600

// RAMRoleArnCredential is a kind of credentials
type RAMRoleArnCredential struct {
	*credentialUpdater
	AccessKeyId           string
	AccessKeySecret       string
	RoleArn               string
	RoleSessionName       string
	RoleSessionExpiration int
	Policy                string
	sessionCredential     *sessionCredential
	runtime               *utils.Runtime
}

type ramRoleArnResponse struct {
	Credentials *credentialsInResponse `json:"Credentials" xml:"Credentials"`
}

type credentialsInResponse struct {
	AccessKeyId     string `json:"AccessKeyId" xml:"AccessKeyId"`
	AccessKeySecret string `json:"AccessKeySecret" xml:"AccessKeySecret"`
	SecurityToken   string `json:"SecurityToken" xml:"SecurityToken"`
	Expiration      string `json:"Expiration" xml:"Expiration"`
}

func newRAMRoleArnCredential(accessKeyId, accessKeySecret, roleArn, roleSessionName, policy string, roleSessionExpiration int, runtime *utils.Runtime) *RAMRoleArnCredential {
	return &RAMRoleArnCredential{
		AccessKeyId:           accessKeyId,
		AccessKeySecret:       accessKeySecret,
		RoleArn:               roleArn,
		RoleSessionName:       roleSessionName,
		RoleSessionExpiration: roleSessionExpiration,
		Policy:                policy,
		credentialUpdater:     new(credentialUpdater),
		runtime:               runtime,
	}
}

// GetAccessKeyId reutrns RamRoleArnCredential's AccessKeyId
// if AccessKeyId is not exist or out of date, the function will update it.
func (r *RAMRoleArnCredential) GetAccessKeyId() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.AccessKeyId), nil
}

// GetAccessSecret reutrns RamRoleArnCredential's AccessKeySecret
// if AccessKeySecret is not exist or out of date, the function will update it.
func (r *RAMRoleArnCredential) GetAccessKeySecret() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.AccessKeySecret), nil
}

// GetSecurityToken reutrns RamRoleArnCredential's SecurityToken
// if SecurityToken is not exist or out of date, the function will update it.
func (r *RAMRoleArnCredential) GetSecurityToken() (*string, error) {
	if r.sessionCredential == nil || r.needUpdateCredential() {
		err := r.updateCredential()
		if err != nil {
			return tea.String(""), err
		}
	}
	return tea.String(r.sessionCredential.SecurityToken), nil
}

// GetBearerToken is useless RamRoleArnCredential
func (r *RAMRoleArnCredential) GetBearerToken() *string {
	return tea.String("")
}

// GetType reutrns RamRoleArnCredential's type
func (r *RAMRoleArnCredential) GetType() *string {
	return tea.String("ram_role_arn")
}

func (r *RAMRoleArnCredential) updateCredential() (err error) {
	if r.runtime == nil {
		r.runtime = new(utils.Runtime)
	}
	request := request.NewCommonRequest()
	request.Domain = "sts.aliyuncs.com"
	request.Scheme = "HTTPS"
	request.Method = "GET"
	request.QueryParams["AccessKeyId"] = r.AccessKeyId
	request.QueryParams["Action"] = "AssumeRole"
	request.QueryParams["Format"] = "JSON"
	if r.RoleSessionExpiration > 0 {
		if r.RoleSessionExpiration >= 900 && r.RoleSessionExpiration <= 3600 {
			request.QueryParams["DurationSeconds"] = strconv.Itoa(r.RoleSessionExpiration)
		} else {
			err = errors.New("[InvalidParam]:Assume Role session duration should be in the range of 15min - 1Hr")
			return
		}
	} else {
		request.QueryParams["DurationSeconds"] = strconv.Itoa(defaultDurationSeconds)
	}
	request.QueryParams["RoleArn"] = r.RoleArn
	if r.Policy != "" {
		request.QueryParams["Policy"] = r.Policy
	}
	request.QueryParams["RoleSessionName"] = r.RoleSessionName
	request.QueryParams["SignatureMethod"] = "HMAC-SHA1"
	request.QueryParams["SignatureVersion"] = "1.0"
	request.QueryParams["Version"] = "2015-04-01"
	request.QueryParams["Timestamp"] = utils.GetTimeInFormatISO8601()
	request.QueryParams["SignatureNonce"] = utils.GetUUID()
	signature := utils.ShaHmac1(request.BuildStringToSign(), r.AccessKeySecret+"&")
	request.QueryParams["Signature"] = signature
	request.Headers["Host"] = request.Domain
	request.Headers["Accept-Encoding"] = "identity"
	request.URL = request.BuildURL()
	content, err := doAction(request, r.runtime)
	if err != nil {
		return fmt.Errorf("refresh RoleArn sts token err: %s", err.Error())
	}
	var resp *ramRoleArnResponse
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
