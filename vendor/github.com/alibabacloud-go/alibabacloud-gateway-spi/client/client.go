// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"io"

	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type InterceptorContext struct {
	Request       *InterceptorContextRequest       `json:"request,omitempty" xml:"request,omitempty" require:"true" type:"Struct"`
	Configuration *InterceptorContextConfiguration `json:"configuration,omitempty" xml:"configuration,omitempty" require:"true" type:"Struct"`
	Response      *InterceptorContextResponse      `json:"response,omitempty" xml:"response,omitempty" require:"true" type:"Struct"`
}

func (s InterceptorContext) String() string {
	return tea.Prettify(s)
}

func (s InterceptorContext) GoString() string {
	return s.String()
}

func (s *InterceptorContext) SetRequest(v *InterceptorContextRequest) *InterceptorContext {
	s.Request = v
	return s
}

func (s *InterceptorContext) SetConfiguration(v *InterceptorContextConfiguration) *InterceptorContext {
	s.Configuration = v
	return s
}

func (s *InterceptorContext) SetResponse(v *InterceptorContextResponse) *InterceptorContext {
	s.Response = v
	return s
}

type InterceptorContextRequest struct {
	Headers            map[string]*string    `json:"headers,omitempty" xml:"headers,omitempty"`
	Query              map[string]*string    `json:"query,omitempty" xml:"query,omitempty"`
	Body               interface{}           `json:"body,omitempty" xml:"body,omitempty"`
	Stream             io.Reader             `json:"stream,omitempty" xml:"stream,omitempty"`
	HostMap            map[string]*string    `json:"hostMap,omitempty" xml:"hostMap,omitempty"`
	Pathname           *string               `json:"pathname,omitempty" xml:"pathname,omitempty" require:"true"`
	ProductId          *string               `json:"productId,omitempty" xml:"productId,omitempty" require:"true"`
	Action             *string               `json:"action,omitempty" xml:"action,omitempty" require:"true"`
	Version            *string               `json:"version,omitempty" xml:"version,omitempty" require:"true"`
	Protocol           *string               `json:"protocol,omitempty" xml:"protocol,omitempty" require:"true"`
	Method             *string               `json:"method,omitempty" xml:"method,omitempty" require:"true"`
	AuthType           *string               `json:"authType,omitempty" xml:"authType,omitempty" require:"true"`
	BodyType           *string               `json:"bodyType,omitempty" xml:"bodyType,omitempty" require:"true"`
	ReqBodyType        *string               `json:"reqBodyType,omitempty" xml:"reqBodyType,omitempty" require:"true"`
	Style              *string               `json:"style,omitempty" xml:"style,omitempty"`
	Credential         credential.Credential `json:"credential,omitempty" xml:"credential,omitempty" require:"true"`
	SignatureVersion   *string               `json:"signatureVersion,omitempty" xml:"signatureVersion,omitempty"`
	SignatureAlgorithm *string               `json:"signatureAlgorithm,omitempty" xml:"signatureAlgorithm,omitempty"`
	UserAgent          *string               `json:"userAgent,omitempty" xml:"userAgent,omitempty" require:"true"`
}

func (s InterceptorContextRequest) String() string {
	return tea.Prettify(s)
}

func (s InterceptorContextRequest) GoString() string {
	return s.String()
}

func (s *InterceptorContextRequest) SetHeaders(v map[string]*string) *InterceptorContextRequest {
	s.Headers = v
	return s
}

func (s *InterceptorContextRequest) SetQuery(v map[string]*string) *InterceptorContextRequest {
	s.Query = v
	return s
}

func (s *InterceptorContextRequest) SetBody(v interface{}) *InterceptorContextRequest {
	s.Body = v
	return s
}

func (s *InterceptorContextRequest) SetStream(v io.Reader) *InterceptorContextRequest {
	s.Stream = v
	return s
}

func (s *InterceptorContextRequest) SetHostMap(v map[string]*string) *InterceptorContextRequest {
	s.HostMap = v
	return s
}

func (s *InterceptorContextRequest) SetPathname(v string) *InterceptorContextRequest {
	s.Pathname = &v
	return s
}

func (s *InterceptorContextRequest) SetProductId(v string) *InterceptorContextRequest {
	s.ProductId = &v
	return s
}

func (s *InterceptorContextRequest) SetAction(v string) *InterceptorContextRequest {
	s.Action = &v
	return s
}

func (s *InterceptorContextRequest) SetVersion(v string) *InterceptorContextRequest {
	s.Version = &v
	return s
}

func (s *InterceptorContextRequest) SetProtocol(v string) *InterceptorContextRequest {
	s.Protocol = &v
	return s
}

func (s *InterceptorContextRequest) SetMethod(v string) *InterceptorContextRequest {
	s.Method = &v
	return s
}

func (s *InterceptorContextRequest) SetAuthType(v string) *InterceptorContextRequest {
	s.AuthType = &v
	return s
}

func (s *InterceptorContextRequest) SetBodyType(v string) *InterceptorContextRequest {
	s.BodyType = &v
	return s
}

func (s *InterceptorContextRequest) SetReqBodyType(v string) *InterceptorContextRequest {
	s.ReqBodyType = &v
	return s
}

func (s *InterceptorContextRequest) SetStyle(v string) *InterceptorContextRequest {
	s.Style = &v
	return s
}

func (s *InterceptorContextRequest) SetCredential(v credential.Credential) *InterceptorContextRequest {
	s.Credential = v
	return s
}

func (s *InterceptorContextRequest) SetSignatureVersion(v string) *InterceptorContextRequest {
	s.SignatureVersion = &v
	return s
}

func (s *InterceptorContextRequest) SetSignatureAlgorithm(v string) *InterceptorContextRequest {
	s.SignatureAlgorithm = &v
	return s
}

func (s *InterceptorContextRequest) SetUserAgent(v string) *InterceptorContextRequest {
	s.UserAgent = &v
	return s
}

type InterceptorContextConfiguration struct {
	RegionId     *string            `json:"regionId,omitempty" xml:"regionId,omitempty" require:"true"`
	Endpoint     *string            `json:"endpoint,omitempty" xml:"endpoint,omitempty"`
	EndpointRule *string            `json:"endpointRule,omitempty" xml:"endpointRule,omitempty"`
	EndpointMap  map[string]*string `json:"endpointMap,omitempty" xml:"endpointMap,omitempty"`
	EndpointType *string            `json:"endpointType,omitempty" xml:"endpointType,omitempty"`
	Network      *string            `json:"network,omitempty" xml:"network,omitempty"`
	Suffix       *string            `json:"suffix,omitempty" xml:"suffix,omitempty"`
}

func (s InterceptorContextConfiguration) String() string {
	return tea.Prettify(s)
}

func (s InterceptorContextConfiguration) GoString() string {
	return s.String()
}

func (s *InterceptorContextConfiguration) SetRegionId(v string) *InterceptorContextConfiguration {
	s.RegionId = &v
	return s
}

func (s *InterceptorContextConfiguration) SetEndpoint(v string) *InterceptorContextConfiguration {
	s.Endpoint = &v
	return s
}

func (s *InterceptorContextConfiguration) SetEndpointRule(v string) *InterceptorContextConfiguration {
	s.EndpointRule = &v
	return s
}

func (s *InterceptorContextConfiguration) SetEndpointMap(v map[string]*string) *InterceptorContextConfiguration {
	s.EndpointMap = v
	return s
}

func (s *InterceptorContextConfiguration) SetEndpointType(v string) *InterceptorContextConfiguration {
	s.EndpointType = &v
	return s
}

func (s *InterceptorContextConfiguration) SetNetwork(v string) *InterceptorContextConfiguration {
	s.Network = &v
	return s
}

func (s *InterceptorContextConfiguration) SetSuffix(v string) *InterceptorContextConfiguration {
	s.Suffix = &v
	return s
}

type InterceptorContextResponse struct {
	StatusCode       *int               `json:"statusCode,omitempty" xml:"statusCode,omitempty"`
	Headers          map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	Body             io.Reader          `json:"body,omitempty" xml:"body,omitempty"`
	DeserializedBody interface{}        `json:"deserializedBody,omitempty" xml:"deserializedBody,omitempty"`
}

func (s InterceptorContextResponse) String() string {
	return tea.Prettify(s)
}

func (s InterceptorContextResponse) GoString() string {
	return s.String()
}

func (s *InterceptorContextResponse) SetStatusCode(v int) *InterceptorContextResponse {
	s.StatusCode = &v
	return s
}

func (s *InterceptorContextResponse) SetHeaders(v map[string]*string) *InterceptorContextResponse {
	s.Headers = v
	return s
}

func (s *InterceptorContextResponse) SetBody(v io.Reader) *InterceptorContextResponse {
	s.Body = v
	return s
}

func (s *InterceptorContextResponse) SetDeserializedBody(v interface{}) *InterceptorContextResponse {
	s.DeserializedBody = v
	return s
}

type AttributeMap struct {
	Attributes map[string]interface{} `json:"attributes,omitempty" xml:"attributes,omitempty" require:"true"`
	Key        map[string]*string     `json:"key,omitempty" xml:"key,omitempty" require:"true"`
}

func (s AttributeMap) String() string {
	return tea.Prettify(s)
}

func (s AttributeMap) GoString() string {
	return s.String()
}

func (s *AttributeMap) SetAttributes(v map[string]interface{}) *AttributeMap {
	s.Attributes = v
	return s
}

func (s *AttributeMap) SetKey(v map[string]*string) *AttributeMap {
	s.Key = v
	return s
}

type ClientInterface interface {
	ModifyConfiguration(context *InterceptorContext, attributeMap *AttributeMap) error
	ModifyRequest(context *InterceptorContext, attributeMap *AttributeMap) error
	ModifyResponse(context *InterceptorContext, attributeMap *AttributeMap) error
}

type Client struct {
}

func NewClient() (*Client, error) {
	client := new(Client)
	err := client.Init()
	return client, err
}

func (client *Client) Init() (_err error) {
	return nil
}

func (client *Client) ModifyConfiguration(context *InterceptorContext, attributeMap *AttributeMap) (_err error) {
	panic("No Support!")
}

func (client *Client) ModifyRequest(context *InterceptorContext, attributeMap *AttributeMap) (_err error) {
	panic("No Support!")
}

func (client *Client) ModifyResponse(context *InterceptorContext, attributeMap *AttributeMap) (_err error) {
	panic("No Support!")
}
