// This file is auto-generated, don't edit it. Thanks.
/**
 * This is for OpenApi SDK
 */
package client

import (
	"io"

	spi "github.com/alibabacloud-go/alibabacloud-gateway-spi/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/service"
	xml "github.com/alibabacloud-go/tea-xml/service"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type GlobalParameters struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	Queries map[string]*string `json:"queries,omitempty" xml:"queries,omitempty"`
}

func (s GlobalParameters) String() string {
	return tea.Prettify(s)
}

func (s GlobalParameters) GoString() string {
	return s.String()
}

func (s *GlobalParameters) SetHeaders(v map[string]*string) *GlobalParameters {
	s.Headers = v
	return s
}

func (s *GlobalParameters) SetQueries(v map[string]*string) *GlobalParameters {
	s.Queries = v
	return s
}

/**
 * Model for initing client
 */
type Config struct {
	// accesskey id
	AccessKeyId *string `json:"accessKeyId,omitempty" xml:"accessKeyId,omitempty"`
	// accesskey secret
	AccessKeySecret *string `json:"accessKeySecret,omitempty" xml:"accessKeySecret,omitempty"`
	// security token
	SecurityToken *string `json:"securityToken,omitempty" xml:"securityToken,omitempty"`
	// http protocol
	Protocol *string `json:"protocol,omitempty" xml:"protocol,omitempty"`
	// http method
	Method *string `json:"method,omitempty" xml:"method,omitempty"`
	// region id
	RegionId *string `json:"regionId,omitempty" xml:"regionId,omitempty"`
	// read timeout
	ReadTimeout *int `json:"readTimeout,omitempty" xml:"readTimeout,omitempty"`
	// connect timeout
	ConnectTimeout *int `json:"connectTimeout,omitempty" xml:"connectTimeout,omitempty"`
	// http proxy
	HttpProxy *string `json:"httpProxy,omitempty" xml:"httpProxy,omitempty"`
	// https proxy
	HttpsProxy *string `json:"httpsProxy,omitempty" xml:"httpsProxy,omitempty"`
	// credential
	Credential credential.Credential `json:"credential,omitempty" xml:"credential,omitempty"`
	// endpoint
	Endpoint *string `json:"endpoint,omitempty" xml:"endpoint,omitempty"`
	// proxy white list
	NoProxy *string `json:"noProxy,omitempty" xml:"noProxy,omitempty"`
	// max idle conns
	MaxIdleConns *int `json:"maxIdleConns,omitempty" xml:"maxIdleConns,omitempty"`
	// network for endpoint
	Network *string `json:"network,omitempty" xml:"network,omitempty"`
	// user agent
	UserAgent *string `json:"userAgent,omitempty" xml:"userAgent,omitempty"`
	// suffix for endpoint
	Suffix *string `json:"suffix,omitempty" xml:"suffix,omitempty"`
	// socks5 proxy
	Socks5Proxy *string `json:"socks5Proxy,omitempty" xml:"socks5Proxy,omitempty"`
	// socks5 network
	Socks5NetWork *string `json:"socks5NetWork,omitempty" xml:"socks5NetWork,omitempty"`
	// endpoint type
	EndpointType *string `json:"endpointType,omitempty" xml:"endpointType,omitempty"`
	// OpenPlatform endpoint
	OpenPlatformEndpoint *string `json:"openPlatformEndpoint,omitempty" xml:"openPlatformEndpoint,omitempty"`
	// Deprecated
	// credential type
	Type *string `json:"type,omitempty" xml:"type,omitempty"`
	// Signature Version
	SignatureVersion *string `json:"signatureVersion,omitempty" xml:"signatureVersion,omitempty"`
	// Signature Algorithm
	SignatureAlgorithm *string `json:"signatureAlgorithm,omitempty" xml:"signatureAlgorithm,omitempty"`
	// Global Parameters
	GlobalParameters *GlobalParameters `json:"globalParameters,omitempty" xml:"globalParameters,omitempty"`
}

func (s Config) String() string {
	return tea.Prettify(s)
}

func (s Config) GoString() string {
	return s.String()
}

func (s *Config) SetAccessKeyId(v string) *Config {
	s.AccessKeyId = &v
	return s
}

func (s *Config) SetAccessKeySecret(v string) *Config {
	s.AccessKeySecret = &v
	return s
}

func (s *Config) SetSecurityToken(v string) *Config {
	s.SecurityToken = &v
	return s
}

func (s *Config) SetProtocol(v string) *Config {
	s.Protocol = &v
	return s
}

func (s *Config) SetMethod(v string) *Config {
	s.Method = &v
	return s
}

func (s *Config) SetRegionId(v string) *Config {
	s.RegionId = &v
	return s
}

func (s *Config) SetReadTimeout(v int) *Config {
	s.ReadTimeout = &v
	return s
}

func (s *Config) SetConnectTimeout(v int) *Config {
	s.ConnectTimeout = &v
	return s
}

func (s *Config) SetHttpProxy(v string) *Config {
	s.HttpProxy = &v
	return s
}

func (s *Config) SetHttpsProxy(v string) *Config {
	s.HttpsProxy = &v
	return s
}

func (s *Config) SetCredential(v credential.Credential) *Config {
	s.Credential = v
	return s
}

func (s *Config) SetEndpoint(v string) *Config {
	s.Endpoint = &v
	return s
}

func (s *Config) SetNoProxy(v string) *Config {
	s.NoProxy = &v
	return s
}

func (s *Config) SetMaxIdleConns(v int) *Config {
	s.MaxIdleConns = &v
	return s
}

func (s *Config) SetNetwork(v string) *Config {
	s.Network = &v
	return s
}

func (s *Config) SetUserAgent(v string) *Config {
	s.UserAgent = &v
	return s
}

func (s *Config) SetSuffix(v string) *Config {
	s.Suffix = &v
	return s
}

func (s *Config) SetSocks5Proxy(v string) *Config {
	s.Socks5Proxy = &v
	return s
}

func (s *Config) SetSocks5NetWork(v string) *Config {
	s.Socks5NetWork = &v
	return s
}

func (s *Config) SetEndpointType(v string) *Config {
	s.EndpointType = &v
	return s
}

func (s *Config) SetOpenPlatformEndpoint(v string) *Config {
	s.OpenPlatformEndpoint = &v
	return s
}

func (s *Config) SetType(v string) *Config {
	s.Type = &v
	return s
}

func (s *Config) SetSignatureVersion(v string) *Config {
	s.SignatureVersion = &v
	return s
}

func (s *Config) SetSignatureAlgorithm(v string) *Config {
	s.SignatureAlgorithm = &v
	return s
}

func (s *Config) SetGlobalParameters(v *GlobalParameters) *Config {
	s.GlobalParameters = v
	return s
}

type OpenApiRequest struct {
	Headers          map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	Query            map[string]*string `json:"query,omitempty" xml:"query,omitempty"`
	Body             interface{}        `json:"body,omitempty" xml:"body,omitempty"`
	Stream           io.Reader          `json:"stream,omitempty" xml:"stream,omitempty"`
	HostMap          map[string]*string `json:"hostMap,omitempty" xml:"hostMap,omitempty"`
	EndpointOverride *string            `json:"endpointOverride,omitempty" xml:"endpointOverride,omitempty"`
}

func (s OpenApiRequest) String() string {
	return tea.Prettify(s)
}

func (s OpenApiRequest) GoString() string {
	return s.String()
}

func (s *OpenApiRequest) SetHeaders(v map[string]*string) *OpenApiRequest {
	s.Headers = v
	return s
}

func (s *OpenApiRequest) SetQuery(v map[string]*string) *OpenApiRequest {
	s.Query = v
	return s
}

func (s *OpenApiRequest) SetBody(v interface{}) *OpenApiRequest {
	s.Body = v
	return s
}

func (s *OpenApiRequest) SetStream(v io.Reader) *OpenApiRequest {
	s.Stream = v
	return s
}

func (s *OpenApiRequest) SetHostMap(v map[string]*string) *OpenApiRequest {
	s.HostMap = v
	return s
}

func (s *OpenApiRequest) SetEndpointOverride(v string) *OpenApiRequest {
	s.EndpointOverride = &v
	return s
}

type Params struct {
	Action      *string `json:"action,omitempty" xml:"action,omitempty" require:"true"`
	Version     *string `json:"version,omitempty" xml:"version,omitempty" require:"true"`
	Protocol    *string `json:"protocol,omitempty" xml:"protocol,omitempty" require:"true"`
	Pathname    *string `json:"pathname,omitempty" xml:"pathname,omitempty" require:"true"`
	Method      *string `json:"method,omitempty" xml:"method,omitempty" require:"true"`
	AuthType    *string `json:"authType,omitempty" xml:"authType,omitempty" require:"true"`
	BodyType    *string `json:"bodyType,omitempty" xml:"bodyType,omitempty" require:"true"`
	ReqBodyType *string `json:"reqBodyType,omitempty" xml:"reqBodyType,omitempty" require:"true"`
	Style       *string `json:"style,omitempty" xml:"style,omitempty"`
}

func (s Params) String() string {
	return tea.Prettify(s)
}

func (s Params) GoString() string {
	return s.String()
}

func (s *Params) SetAction(v string) *Params {
	s.Action = &v
	return s
}

func (s *Params) SetVersion(v string) *Params {
	s.Version = &v
	return s
}

func (s *Params) SetProtocol(v string) *Params {
	s.Protocol = &v
	return s
}

func (s *Params) SetPathname(v string) *Params {
	s.Pathname = &v
	return s
}

func (s *Params) SetMethod(v string) *Params {
	s.Method = &v
	return s
}

func (s *Params) SetAuthType(v string) *Params {
	s.AuthType = &v
	return s
}

func (s *Params) SetBodyType(v string) *Params {
	s.BodyType = &v
	return s
}

func (s *Params) SetReqBodyType(v string) *Params {
	s.ReqBodyType = &v
	return s
}

func (s *Params) SetStyle(v string) *Params {
	s.Style = &v
	return s
}

type Client struct {
	Endpoint             *string
	RegionId             *string
	Protocol             *string
	Method               *string
	UserAgent            *string
	EndpointRule         *string
	EndpointMap          map[string]*string
	Suffix               *string
	ReadTimeout          *int
	ConnectTimeout       *int
	HttpProxy            *string
	HttpsProxy           *string
	Socks5Proxy          *string
	Socks5NetWork        *string
	NoProxy              *string
	Network              *string
	ProductId            *string
	MaxIdleConns         *int
	EndpointType         *string
	OpenPlatformEndpoint *string
	Credential           credential.Credential
	SignatureVersion     *string
	SignatureAlgorithm   *string
	Headers              map[string]*string
	Spi                  spi.ClientInterface
	GlobalParameters     *GlobalParameters
}

/**
 * Init client with Config
 * @param config config contains the necessary information to create a client
 */
func NewClient(config *Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *Config) (_err error) {
	if tea.BoolValue(util.IsUnset(tea.ToMap(config))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'config' can not be unset",
		})
		return _err
	}

	if !tea.BoolValue(util.Empty(config.AccessKeyId)) && !tea.BoolValue(util.Empty(config.AccessKeySecret)) {
		if !tea.BoolValue(util.Empty(config.SecurityToken)) {
			config.Type = tea.String("sts")
		} else {
			config.Type = tea.String("access_key")
		}

		credentialConfig := &credential.Config{
			AccessKeyId:     config.AccessKeyId,
			Type:            config.Type,
			AccessKeySecret: config.AccessKeySecret,
			SecurityToken:   config.SecurityToken,
		}
		client.Credential, _err = credential.NewCredential(credentialConfig)
		if _err != nil {
			return _err
		}

	} else if !tea.BoolValue(util.IsUnset(config.Credential)) {
		client.Credential = config.Credential
	}

	client.Endpoint = config.Endpoint
	client.EndpointType = config.EndpointType
	client.Network = config.Network
	client.Suffix = config.Suffix
	client.Protocol = config.Protocol
	client.Method = config.Method
	client.RegionId = config.RegionId
	client.UserAgent = config.UserAgent
	client.ReadTimeout = config.ReadTimeout
	client.ConnectTimeout = config.ConnectTimeout
	client.HttpProxy = config.HttpProxy
	client.HttpsProxy = config.HttpsProxy
	client.NoProxy = config.NoProxy
	client.Socks5Proxy = config.Socks5Proxy
	client.Socks5NetWork = config.Socks5NetWork
	client.MaxIdleConns = config.MaxIdleConns
	client.SignatureVersion = config.SignatureVersion
	client.SignatureAlgorithm = config.SignatureAlgorithm
	client.GlobalParameters = config.GlobalParameters
	return nil
}

/**
 * Encapsulate the request and invoke the network
 * @param action api name
 * @param version product version
 * @param protocol http or https
 * @param method e.g. GET
 * @param authType authorization type e.g. AK
 * @param bodyType response body type e.g. String
 * @param request object of OpenApiRequest
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) DoRPCRequest(action *string, version *string, protocol *string, method *string, authType *string, bodyType *string, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(request)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"socks5Proxy":    tea.StringValue(util.DefaultString(runtime.Socks5Proxy, client.Socks5Proxy)),
		"socks5NetWork":  tea.StringValue(util.DefaultString(runtime.Socks5NetWork, client.Socks5NetWork)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, protocol)
			request_.Method = method
			request_.Pathname = tea.String("/")
			request_.Query = tea.Merge(map[string]*string{
				"Action":         action,
				"Format":         tea.String("json"),
				"Version":        version,
				"Timestamp":      openapiutil.GetTimestamp(),
				"SignatureNonce": util.GetNonce(),
			}, request.Query)
			headers, _err := client.GetRpcHeaders()
			if _err != nil {
				return _result, _err
			}

			if tea.BoolValue(util.IsUnset(headers)) {
				// endpoint is setted in product client
				request_.Headers = map[string]*string{
					"host":          client.Endpoint,
					"x-acs-version": version,
					"x-acs-action":  action,
					"user-agent":    client.GetUserAgent(),
				}
			} else {
				request_.Headers = tea.Merge(map[string]*string{
					"host":          client.Endpoint,
					"x-acs-version": version,
					"x-acs-action":  action,
					"user-agent":    client.GetUserAgent(),
				}, headers)
			}

			if !tea.BoolValue(util.IsUnset(request.Body)) {
				m := util.AssertAsMap(request.Body)
				tmp := util.AnyifyMapValue(openapiutil.Query(m))
				request_.Body = tea.ToReader(util.ToFormString(tmp))
				request_.Headers["content-type"] = tea.String("application/x-www-form-urlencoded")
			}

			if !tea.BoolValue(util.EqualString(authType, tea.String("Anonymous"))) {
				accessKeyId, _err := client.GetAccessKeyId()
				if _err != nil {
					return _result, _err
				}

				accessKeySecret, _err := client.GetAccessKeySecret()
				if _err != nil {
					return _result, _err
				}

				securityToken, _err := client.GetSecurityToken()
				if _err != nil {
					return _result, _err
				}

				if !tea.BoolValue(util.Empty(securityToken)) {
					request_.Query["SecurityToken"] = securityToken
				}

				request_.Query["SignatureMethod"] = tea.String("HMAC-SHA1")
				request_.Query["SignatureVersion"] = tea.String("1.0")
				request_.Query["AccessKeyId"] = accessKeyId
				var t map[string]interface{}
				if !tea.BoolValue(util.IsUnset(request.Body)) {
					t = util.AssertAsMap(request.Body)
				}

				signedParam := tea.Merge(request_.Query,
					openapiutil.Query(t))
				request_.Query["Signature"] = openapiutil.GetRPCSignature(signedParam, request_.Method, accessKeySecret)
			}

			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				_res, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				err := util.AssertAsMap(_res)
				requestId := DefaultAny(err["RequestId"], err["requestId"])
				_err = tea.NewSDKError(map[string]interface{}{
					"code":    tea.ToString(DefaultAny(err["Code"], err["code"])),
					"message": "code: " + tea.ToString(tea.IntValue(response_.StatusCode)) + ", " + tea.ToString(DefaultAny(err["Message"], err["message"])) + " request id: " + tea.ToString(requestId),
					"data":    err,
				})
				return _result, _err
			}

			if tea.BoolValue(util.EqualString(bodyType, tea.String("binary"))) {
				resp := map[string]interface{}{
					"body":    response_.Body,
					"headers": response_.Headers,
				}
				_result = resp
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("byte"))) {
				byt, _err := util.ReadAsBytes(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    byt,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("string"))) {
				str, _err := util.ReadAsString(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    tea.StringValue(str),
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("json"))) {
				obj, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				res := util.AssertAsMap(obj)
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    res,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("array"))) {
				arr, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    arr,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]map[string]*string{
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * Encapsulate the request and invoke the network
 * @param action api name
 * @param version product version
 * @param protocol http or https
 * @param method e.g. GET
 * @param authType authorization type e.g. AK
 * @param pathname pathname of every api
 * @param bodyType response body type e.g. String
 * @param request object of OpenApiRequest
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) DoROARequest(action *string, version *string, protocol *string, method *string, authType *string, pathname *string, bodyType *string, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(request)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"socks5Proxy":    tea.StringValue(util.DefaultString(runtime.Socks5Proxy, client.Socks5Proxy)),
		"socks5NetWork":  tea.StringValue(util.DefaultString(runtime.Socks5NetWork, client.Socks5NetWork)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, protocol)
			request_.Method = method
			request_.Pathname = pathname
			request_.Headers = tea.Merge(map[string]*string{
				"date":                    util.GetDateUTCString(),
				"host":                    client.Endpoint,
				"accept":                  tea.String("application/json"),
				"x-acs-signature-nonce":   util.GetNonce(),
				"x-acs-signature-method":  tea.String("HMAC-SHA1"),
				"x-acs-signature-version": tea.String("1.0"),
				"x-acs-version":           version,
				"x-acs-action":            action,
				"user-agent":              util.GetUserAgent(client.UserAgent),
			}, request.Headers)
			if !tea.BoolValue(util.IsUnset(request.Body)) {
				request_.Body = tea.ToReader(util.ToJSONString(request.Body))
				request_.Headers["content-type"] = tea.String("application/json; charset=utf-8")
			}

			if !tea.BoolValue(util.IsUnset(request.Query)) {
				request_.Query = request.Query
			}

			if !tea.BoolValue(util.EqualString(authType, tea.String("Anonymous"))) {
				accessKeyId, _err := client.GetAccessKeyId()
				if _err != nil {
					return _result, _err
				}

				accessKeySecret, _err := client.GetAccessKeySecret()
				if _err != nil {
					return _result, _err
				}

				securityToken, _err := client.GetSecurityToken()
				if _err != nil {
					return _result, _err
				}

				if !tea.BoolValue(util.Empty(securityToken)) {
					request_.Headers["x-acs-accesskey-id"] = accessKeyId
					request_.Headers["x-acs-security-token"] = securityToken
				}

				stringToSign := openapiutil.GetStringToSign(request_)
				request_.Headers["authorization"] = tea.String("acs " + tea.StringValue(accessKeyId) + ":" + tea.StringValue(openapiutil.GetROASignature(stringToSign, accessKeySecret)))
			}

			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			if tea.BoolValue(util.EqualNumber(response_.StatusCode, tea.Int(204))) {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]map[string]*string{
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				_res, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				err := util.AssertAsMap(_res)
				requestId := DefaultAny(err["RequestId"], err["requestId"])
				requestId = DefaultAny(requestId, err["requestid"])
				_err = tea.NewSDKError(map[string]interface{}{
					"code":    tea.ToString(DefaultAny(err["Code"], err["code"])),
					"message": "code: " + tea.ToString(tea.IntValue(response_.StatusCode)) + ", " + tea.ToString(DefaultAny(err["Message"], err["message"])) + " request id: " + tea.ToString(requestId),
					"data":    err,
				})
				return _result, _err
			}

			if tea.BoolValue(util.EqualString(bodyType, tea.String("binary"))) {
				resp := map[string]interface{}{
					"body":    response_.Body,
					"headers": response_.Headers,
				}
				_result = resp
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("byte"))) {
				byt, _err := util.ReadAsBytes(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    byt,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("string"))) {
				str, _err := util.ReadAsString(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    tea.StringValue(str),
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("json"))) {
				obj, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				res := util.AssertAsMap(obj)
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    res,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("array"))) {
				arr, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    arr,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]map[string]*string{
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * Encapsulate the request and invoke the network with form body
 * @param action api name
 * @param version product version
 * @param protocol http or https
 * @param method e.g. GET
 * @param authType authorization type e.g. AK
 * @param pathname pathname of every api
 * @param bodyType response body type e.g. String
 * @param request object of OpenApiRequest
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) DoROARequestWithForm(action *string, version *string, protocol *string, method *string, authType *string, pathname *string, bodyType *string, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(request)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"socks5Proxy":    tea.StringValue(util.DefaultString(runtime.Socks5Proxy, client.Socks5Proxy)),
		"socks5NetWork":  tea.StringValue(util.DefaultString(runtime.Socks5NetWork, client.Socks5NetWork)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, protocol)
			request_.Method = method
			request_.Pathname = pathname
			request_.Headers = tea.Merge(map[string]*string{
				"date":                    util.GetDateUTCString(),
				"host":                    client.Endpoint,
				"accept":                  tea.String("application/json"),
				"x-acs-signature-nonce":   util.GetNonce(),
				"x-acs-signature-method":  tea.String("HMAC-SHA1"),
				"x-acs-signature-version": tea.String("1.0"),
				"x-acs-version":           version,
				"x-acs-action":            action,
				"user-agent":              util.GetUserAgent(client.UserAgent),
			}, request.Headers)
			if !tea.BoolValue(util.IsUnset(request.Body)) {
				m := util.AssertAsMap(request.Body)
				request_.Body = tea.ToReader(openapiutil.ToForm(m))
				request_.Headers["content-type"] = tea.String("application/x-www-form-urlencoded")
			}

			if !tea.BoolValue(util.IsUnset(request.Query)) {
				request_.Query = request.Query
			}

			if !tea.BoolValue(util.EqualString(authType, tea.String("Anonymous"))) {
				accessKeyId, _err := client.GetAccessKeyId()
				if _err != nil {
					return _result, _err
				}

				accessKeySecret, _err := client.GetAccessKeySecret()
				if _err != nil {
					return _result, _err
				}

				securityToken, _err := client.GetSecurityToken()
				if _err != nil {
					return _result, _err
				}

				if !tea.BoolValue(util.Empty(securityToken)) {
					request_.Headers["x-acs-accesskey-id"] = accessKeyId
					request_.Headers["x-acs-security-token"] = securityToken
				}

				stringToSign := openapiutil.GetStringToSign(request_)
				request_.Headers["authorization"] = tea.String("acs " + tea.StringValue(accessKeyId) + ":" + tea.StringValue(openapiutil.GetROASignature(stringToSign, accessKeySecret)))
			}

			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			if tea.BoolValue(util.EqualNumber(response_.StatusCode, tea.Int(204))) {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]map[string]*string{
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				_res, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				err := util.AssertAsMap(_res)
				_err = tea.NewSDKError(map[string]interface{}{
					"code":    tea.ToString(DefaultAny(err["Code"], err["code"])),
					"message": "code: " + tea.ToString(tea.IntValue(response_.StatusCode)) + ", " + tea.ToString(DefaultAny(err["Message"], err["message"])) + " request id: " + tea.ToString(DefaultAny(err["RequestId"], err["requestId"])),
					"data":    err,
				})
				return _result, _err
			}

			if tea.BoolValue(util.EqualString(bodyType, tea.String("binary"))) {
				resp := map[string]interface{}{
					"body":    response_.Body,
					"headers": response_.Headers,
				}
				_result = resp
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("byte"))) {
				byt, _err := util.ReadAsBytes(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    byt,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("string"))) {
				str, _err := util.ReadAsString(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    tea.StringValue(str),
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("json"))) {
				obj, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				res := util.AssertAsMap(obj)
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    res,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(bodyType, tea.String("array"))) {
				arr, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    arr,
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			} else {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]map[string]*string{
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * Encapsulate the request and invoke the network
 * @param action api name
 * @param version product version
 * @param protocol http or https
 * @param method e.g. GET
 * @param authType authorization type e.g. AK
 * @param bodyType response body type e.g. String
 * @param request object of OpenApiRequest
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) DoRequest(params *Params, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(params)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(request)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"socks5Proxy":    tea.StringValue(util.DefaultString(runtime.Socks5Proxy, client.Socks5Proxy)),
		"socks5NetWork":  tea.StringValue(util.DefaultString(runtime.Socks5NetWork, client.Socks5NetWork)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, params.Protocol)
			request_.Method = params.Method
			request_.Pathname = params.Pathname
			globalQueries := make(map[string]*string)
			globalHeaders := make(map[string]*string)
			if !tea.BoolValue(util.IsUnset(tea.ToMap(client.GlobalParameters))) {
				globalParams := client.GlobalParameters
				if !tea.BoolValue(util.IsUnset(globalParams.Queries)) {
					globalQueries = globalParams.Queries
				}

				if !tea.BoolValue(util.IsUnset(globalParams.Headers)) {
					globalHeaders = globalParams.Headers
				}

			}

			request_.Query = tea.Merge(globalQueries,
				request.Query)
			// endpoint is setted in product client
			request_.Headers = tea.Merge(map[string]*string{
				"host":                  client.Endpoint,
				"x-acs-version":         params.Version,
				"x-acs-action":          params.Action,
				"user-agent":            client.GetUserAgent(),
				"x-acs-date":            openapiutil.GetTimestamp(),
				"x-acs-signature-nonce": util.GetNonce(),
				"accept":                tea.String("application/json"),
			}, globalHeaders,
				request.Headers)
			if tea.BoolValue(util.EqualString(params.Style, tea.String("RPC"))) {
				headers, _err := client.GetRpcHeaders()
				if _err != nil {
					return _result, _err
				}

				if !tea.BoolValue(util.IsUnset(headers)) {
					request_.Headers = tea.Merge(request_.Headers,
						headers)
				}

			}

			signatureAlgorithm := util.DefaultString(client.SignatureAlgorithm, tea.String("ACS3-HMAC-SHA256"))
			hashedRequestPayload := openapiutil.HexEncode(openapiutil.Hash(util.ToBytes(tea.String("")), signatureAlgorithm))
			if !tea.BoolValue(util.IsUnset(request.Stream)) {
				tmp, _err := util.ReadAsBytes(request.Stream)
				if _err != nil {
					return _result, _err
				}

				hashedRequestPayload = openapiutil.HexEncode(openapiutil.Hash(tmp, signatureAlgorithm))
				request_.Body = tea.ToReader(tmp)
				request_.Headers["content-type"] = tea.String("application/octet-stream")
			} else {
				if !tea.BoolValue(util.IsUnset(request.Body)) {
					if tea.BoolValue(util.EqualString(params.ReqBodyType, tea.String("json"))) {
						jsonObj := util.ToJSONString(request.Body)
						hashedRequestPayload = openapiutil.HexEncode(openapiutil.Hash(util.ToBytes(jsonObj), signatureAlgorithm))
						request_.Body = tea.ToReader(jsonObj)
						request_.Headers["content-type"] = tea.String("application/json; charset=utf-8")
					} else {
						m := util.AssertAsMap(request.Body)
						formObj := openapiutil.ToForm(m)
						hashedRequestPayload = openapiutil.HexEncode(openapiutil.Hash(util.ToBytes(formObj), signatureAlgorithm))
						request_.Body = tea.ToReader(formObj)
						request_.Headers["content-type"] = tea.String("application/x-www-form-urlencoded")
					}

				}

			}

			request_.Headers["x-acs-content-sha256"] = hashedRequestPayload
			if !tea.BoolValue(util.EqualString(params.AuthType, tea.String("Anonymous"))) {
				authType, _err := client.GetType()
				if _err != nil {
					return _result, _err
				}

				if tea.BoolValue(util.EqualString(authType, tea.String("bearer"))) {
					bearerToken, _err := client.GetBearerToken()
					if _err != nil {
						return _result, _err
					}

					request_.Headers["x-acs-bearer-token"] = bearerToken
				} else {
					accessKeyId, _err := client.GetAccessKeyId()
					if _err != nil {
						return _result, _err
					}

					accessKeySecret, _err := client.GetAccessKeySecret()
					if _err != nil {
						return _result, _err
					}

					securityToken, _err := client.GetSecurityToken()
					if _err != nil {
						return _result, _err
					}

					if !tea.BoolValue(util.Empty(securityToken)) {
						request_.Headers["x-acs-accesskey-id"] = accessKeyId
						request_.Headers["x-acs-security-token"] = securityToken
					}

					request_.Headers["Authorization"] = openapiutil.GetAuthorization(request_, signatureAlgorithm, hashedRequestPayload, accessKeyId, accessKeySecret)
				}

			}

			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				err := map[string]interface{}{}
				if !tea.BoolValue(util.IsUnset(response_.Headers["content-type"])) && tea.BoolValue(util.EqualString(response_.Headers["content-type"], tea.String("text/xml;charset=utf-8"))) {
					_str, _err := util.ReadAsString(response_.Body)
					if _err != nil {
						return _result, _err
					}

					respMap := xml.ParseXml(_str, nil)
					err = util.AssertAsMap(respMap["Error"])
				} else {
					_res, _err := util.ReadAsJSON(response_.Body)
					if _err != nil {
						return _result, _err
					}

					err = util.AssertAsMap(_res)
				}

				err["statusCode"] = response_.StatusCode
				_err = tea.NewSDKError(map[string]interface{}{
					"code":    tea.ToString(DefaultAny(err["Code"], err["code"])),
					"message": "code: " + tea.ToString(tea.IntValue(response_.StatusCode)) + ", " + tea.ToString(DefaultAny(err["Message"], err["message"])) + " request id: " + tea.ToString(DefaultAny(err["RequestId"], err["requestId"])),
					"data":    err,
				})
				return _result, _err
			}

			if tea.BoolValue(util.EqualString(params.BodyType, tea.String("binary"))) {
				resp := map[string]interface{}{
					"body":       response_.Body,
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}
				_result = resp
				return _result, _err
			} else if tea.BoolValue(util.EqualString(params.BodyType, tea.String("byte"))) {
				byt, _err := util.ReadAsBytes(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":       byt,
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(params.BodyType, tea.String("string"))) {
				str, _err := util.ReadAsString(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":       tea.StringValue(str),
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(params.BodyType, tea.String("json"))) {
				obj, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				res := util.AssertAsMap(obj)
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":       res,
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}, &_result)
				return _result, _err
			} else if tea.BoolValue(util.EqualString(params.BodyType, tea.String("array"))) {
				arr, _err := util.ReadAsJSON(response_.Body)
				if _err != nil {
					return _result, _err
				}

				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":       arr,
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}, &_result)
				return _result, _err
			} else {
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"headers":    response_.Headers,
					"statusCode": tea.IntValue(response_.StatusCode),
				}, &_result)
				return _result, _err
			}

		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * Encapsulate the request and invoke the network
 * @param action api name
 * @param version product version
 * @param protocol http or https
 * @param method e.g. GET
 * @param authType authorization type e.g. AK
 * @param bodyType response body type e.g. String
 * @param request object of OpenApiRequest
 * @param runtime which controls some details of call api, such as retry times
 * @return the response
 */
func (client *Client) Execute(params *Params, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(params)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(request)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(util.DefaultNumber(runtime.ReadTimeout, client.ReadTimeout)),
		"connectTimeout": tea.IntValue(util.DefaultNumber(runtime.ConnectTimeout, client.ConnectTimeout)),
		"httpProxy":      tea.StringValue(util.DefaultString(runtime.HttpProxy, client.HttpProxy)),
		"httpsProxy":     tea.StringValue(util.DefaultString(runtime.HttpsProxy, client.HttpsProxy)),
		"noProxy":        tea.StringValue(util.DefaultString(runtime.NoProxy, client.NoProxy)),
		"socks5Proxy":    tea.StringValue(util.DefaultString(runtime.Socks5Proxy, client.Socks5Proxy)),
		"socks5NetWork":  tea.StringValue(util.DefaultString(runtime.Socks5NetWork, client.Socks5NetWork)),
		"maxIdleConns":   tea.IntValue(util.DefaultNumber(runtime.MaxIdleConns, client.MaxIdleConns)),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(3))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			// spi = new Gateway();//Gateway implements SPI，这一步在产品 SDK 中实例化
			headers, _err := client.GetRpcHeaders()
			if _err != nil {
				return _result, _err
			}

			globalQueries := make(map[string]*string)
			globalHeaders := make(map[string]*string)
			if !tea.BoolValue(util.IsUnset(tea.ToMap(client.GlobalParameters))) {
				globalParams := client.GlobalParameters
				if !tea.BoolValue(util.IsUnset(globalParams.Queries)) {
					globalQueries = globalParams.Queries
				}

				if !tea.BoolValue(util.IsUnset(globalParams.Headers)) {
					globalHeaders = globalParams.Headers
				}

			}

			requestContext := &spi.InterceptorContextRequest{
				Headers: tea.Merge(globalHeaders,
					request.Headers,
					headers),
				Query: tea.Merge(globalQueries,
					request.Query),
				Body:               request.Body,
				Stream:             request.Stream,
				HostMap:            request.HostMap,
				Pathname:           params.Pathname,
				ProductId:          client.ProductId,
				Action:             params.Action,
				Version:            params.Version,
				Protocol:           util.DefaultString(client.Protocol, params.Protocol),
				Method:             util.DefaultString(client.Method, params.Method),
				AuthType:           params.AuthType,
				BodyType:           params.BodyType,
				ReqBodyType:        params.ReqBodyType,
				Style:              params.Style,
				Credential:         client.Credential,
				SignatureVersion:   client.SignatureVersion,
				SignatureAlgorithm: client.SignatureAlgorithm,
				UserAgent:          client.GetUserAgent(),
			}
			configurationContext := &spi.InterceptorContextConfiguration{
				RegionId:     client.RegionId,
				Endpoint:     util.DefaultString(request.EndpointOverride, client.Endpoint),
				EndpointRule: client.EndpointRule,
				EndpointMap:  client.EndpointMap,
				EndpointType: client.EndpointType,
				Network:      client.Network,
				Suffix:       client.Suffix,
			}
			interceptorContext := &spi.InterceptorContext{
				Request:       requestContext,
				Configuration: configurationContext,
			}
			attributeMap := &spi.AttributeMap{}
			// 1. spi.modifyConfiguration(context: SPI.InterceptorContext, attributeMap: SPI.AttributeMap);
			_err = client.Spi.ModifyConfiguration(interceptorContext, attributeMap)
			if _err != nil {
				return _result, _err
			}
			// 2. spi.modifyRequest(context: SPI.InterceptorContext, attributeMap: SPI.AttributeMap);
			_err = client.Spi.ModifyRequest(interceptorContext, attributeMap)
			if _err != nil {
				return _result, _err
			}
			request_.Protocol = interceptorContext.Request.Protocol
			request_.Method = interceptorContext.Request.Method
			request_.Pathname = interceptorContext.Request.Pathname
			request_.Query = interceptorContext.Request.Query
			request_.Body = interceptorContext.Request.Stream
			request_.Headers = interceptorContext.Request.Headers
			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			responseContext := &spi.InterceptorContextResponse{
				StatusCode: response_.StatusCode,
				Headers:    response_.Headers,
				Body:       response_.Body,
			}
			interceptorContext.Response = responseContext
			// 3. spi.modifyResponse(context: SPI.InterceptorContext, attributeMap: SPI.AttributeMap);
			_err = client.Spi.ModifyResponse(interceptorContext, attributeMap)
			if _err != nil {
				return _result, _err
			}
			_result = make(map[string]interface{})
			_err = tea.Convert(map[string]interface{}{
				"headers":    interceptorContext.Response.Headers,
				"statusCode": tea.IntValue(interceptorContext.Response.StatusCode),
				"body":       interceptorContext.Response.DeserializedBody,
			}, &_result)
			return _result, _err
		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

func (client *Client) CallApi(params *Params, request *OpenApiRequest, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	if tea.BoolValue(util.IsUnset(tea.ToMap(params))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'params' can not be unset",
		})
		return _result, _err
	}

	if tea.BoolValue(util.IsUnset(client.SignatureAlgorithm)) || !tea.BoolValue(util.EqualString(client.SignatureAlgorithm, tea.String("v2"))) {
		_result = make(map[string]interface{})
		_body, _err := client.DoRequest(params, request, runtime)
		if _err != nil {
			return _result, _err
		}
		_result = _body
		return _result, _err
	} else if tea.BoolValue(util.EqualString(params.Style, tea.String("ROA"))) && tea.BoolValue(util.EqualString(params.ReqBodyType, tea.String("json"))) {
		_result = make(map[string]interface{})
		_body, _err := client.DoROARequest(params.Action, params.Version, params.Protocol, params.Method, params.AuthType, params.Pathname, params.BodyType, request, runtime)
		if _err != nil {
			return _result, _err
		}
		_result = _body
		return _result, _err
	} else if tea.BoolValue(util.EqualString(params.Style, tea.String("ROA"))) {
		_result = make(map[string]interface{})
		_body, _err := client.DoROARequestWithForm(params.Action, params.Version, params.Protocol, params.Method, params.AuthType, params.Pathname, params.BodyType, request, runtime)
		if _err != nil {
			return _result, _err
		}
		_result = _body
		return _result, _err
	} else {
		_result = make(map[string]interface{})
		_body, _err := client.DoRPCRequest(params.Action, params.Version, params.Protocol, params.Method, params.AuthType, params.BodyType, request, runtime)
		if _err != nil {
			return _result, _err
		}
		_result = _body
		return _result, _err
	}

}

/**
 * Get user agent
 * @return user agent
 */
func (client *Client) GetUserAgent() (_result *string) {
	userAgent := util.GetUserAgent(client.UserAgent)
	_result = userAgent
	return _result
}

/**
 * Get accesskey id by using credential
 * @return accesskey id
 */
func (client *Client) GetAccessKeyId() (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(client.Credential)) {
		_result = tea.String("")
		return _result, _err
	}

	accessKeyId, _err := client.Credential.GetAccessKeyId()
	if _err != nil {
		return _result, _err
	}

	_result = accessKeyId
	return _result, _err
}

/**
 * Get accesskey secret by using credential
 * @return accesskey secret
 */
func (client *Client) GetAccessKeySecret() (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(client.Credential)) {
		_result = tea.String("")
		return _result, _err
	}

	secret, _err := client.Credential.GetAccessKeySecret()
	if _err != nil {
		return _result, _err
	}

	_result = secret
	return _result, _err
}

/**
 * Get security token by using credential
 * @return security token
 */
func (client *Client) GetSecurityToken() (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(client.Credential)) {
		_result = tea.String("")
		return _result, _err
	}

	token, _err := client.Credential.GetSecurityToken()
	if _err != nil {
		return _result, _err
	}

	_result = token
	return _result, _err
}

/**
 * Get bearer token by credential
 * @return bearer token
 */
func (client *Client) GetBearerToken() (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(client.Credential)) {
		_result = tea.String("")
		return _result, _err
	}

	token := client.Credential.GetBearerToken()
	_result = token
	return _result, _err
}

/**
 * Get credential type by credential
 * @return credential type e.g. access_key
 */
func (client *Client) GetType() (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(client.Credential)) {
		_result = tea.String("")
		return _result, _err
	}

	authType := client.Credential.GetType()
	_result = authType
	return _result, _err
}

/**
 * If inputValue is not null, return it or return defaultValue
 * @param inputValue  users input value
 * @param defaultValue default value
 * @return the final result
 */
func DefaultAny(inputValue interface{}, defaultValue interface{}) (_result interface{}) {
	if tea.BoolValue(util.IsUnset(inputValue)) {
		_result = defaultValue
		return _result
	}

	_result = inputValue
	return _result
}

/**
 * If the endpointRule and config.endpoint are empty, throw error
 * @param config config contains the necessary information to create a client
 */
func (client *Client) CheckConfig(config *Config) (_err error) {
	if tea.BoolValue(util.Empty(client.EndpointRule)) && tea.BoolValue(util.Empty(config.Endpoint)) {
		_err = tea.NewSDKError(map[string]interface{}{
			"code":    "ParameterMissing",
			"message": "'config.endpoint' can not be empty",
		})
		return _err
	}

	return _err
}

/**
 * set RPC header for debug
 * @param headers headers for debug, this header can be used only once.
 */
func (client *Client) SetRpcHeaders(headers map[string]*string) (_err error) {
	client.Headers = headers
	return _err
}

/**
 * get RPC header for debug
 */
func (client *Client) GetRpcHeaders() (_result map[string]*string, _err error) {
	headers := client.Headers
	client.Headers = nil
	_result = headers
	return _result, _err
}
