package acr

import (
	"fmt"
	"time"

	cr2016 "github.com/alibabacloud-go/cr-20160607/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/mozillazg/docker-credential-acr-helper/pkg/version"
)

type personClient struct {
	client *cr2016.Client
}

func newPersonClient(region string) (*personClient, error) {
	cred, err := getOpenapiAuth()
	if err != nil {
		return nil, err
	}
	c := &openapi.Config{
		RegionId:   tea.String(region),
		Credential: cred,
		UserAgent:  tea.String(version.UserAgent()),
	}
	client, err := cr2016.NewClient(c)
	if err != nil {
		return nil, err
	}
	return &personClient{client: client}, nil
}

func (c *personClient) getCredentials() (*Credentials, error) {
	resp, err := c.GetAuthorizationToken()
	if err != nil {
		return nil, err
	}
	if resp.Body == nil || resp.Body.Data == nil {
		return nil, fmt.Errorf("get credentials failed: %s", resp.String())
	}

	exp := tea.Int64Value(resp.Body.Data.ExpireTime) / 1000
	expTime := time.Unix(exp, 0).UTC()
	cred := &Credentials{
		UserName:   tea.StringValue(resp.Body.Data.TempUsername),
		Password:   tea.StringValue(resp.Body.Data.AuthorizationToken),
		ExpireTime: expTime,
	}
	return cred, nil
}

func (c *personClient) GetAuthorizationToken() (_result *getPersonAuthorizationTokenResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &getPersonAuthorizationTokenResponse{}
	_body, _err := c.GetDefaultAuthorizationTokenWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (c *personClient) GetDefaultAuthorizationTokenWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *getPersonAuthorizationTokenResponse, _err error) {
	client := c.client
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetAuthorizationToken"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/tokens"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &getPersonAuthorizationTokenResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

type getPersonAuthorizationTokenResponseBody struct {
	Data *getPersonAuthorizationTokenData `json:"data,omitempty" xml:"data,omitempty"`
}

func (s getPersonAuthorizationTokenResponseBody) String() string {
	return tea.Prettify(s)
}
func (s getPersonAuthorizationTokenResponseBody) GoString() string {
	return s.String()
}

type getPersonAuthorizationTokenData struct {
	AuthorizationToken *string `json:"authorizationToken,omitempty" xml:"authorizationToken,omitempty"`
	ExpireTime         *int64  `json:"expireDate,omitempty" xml:"expireDate,omitempty"`
	TempUsername       *string `json:"tempUserName,omitempty" xml:"tempUserName,omitempty"`
	RequestId          *string `json:"requestId,omitempty" xml:"requestId,omitempty"`
	Code               *string `json:"code,omitempty" xml:"code,omitempty"`
}

type getPersonAuthorizationTokenResponse struct {
	Headers map[string]*string                       `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	Body    *getPersonAuthorizationTokenResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s getPersonAuthorizationTokenResponse) String() string {
	return tea.Prettify(s)
}

func (s getPersonAuthorizationTokenResponse) GoString() string {
	return s.String()
}

func (s *getPersonAuthorizationTokenResponse) SetHeaders(v map[string]*string) *getPersonAuthorizationTokenResponse {
	s.Headers = v
	return s
}
