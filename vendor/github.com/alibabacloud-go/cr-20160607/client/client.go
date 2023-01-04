// This file is auto-generated, don't edit it. Thanks.
/**
 *
 */
package client

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	endpointutil "github.com/alibabacloud-go/endpoint-util/service"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

type CancelRepoBuildResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CancelRepoBuildResponse) String() string {
	return tea.Prettify(s)
}

func (s CancelRepoBuildResponse) GoString() string {
	return s.String()
}

func (s *CancelRepoBuildResponse) SetHeaders(v map[string]*string) *CancelRepoBuildResponse {
	s.Headers = v
	return s
}

type CreateNamespaceResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CreateNamespaceResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateNamespaceResponse) GoString() string {
	return s.String()
}

func (s *CreateNamespaceResponse) SetHeaders(v map[string]*string) *CreateNamespaceResponse {
	s.Headers = v
	return s
}

type CreateRepoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CreateRepoResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateRepoResponse) GoString() string {
	return s.String()
}

func (s *CreateRepoResponse) SetHeaders(v map[string]*string) *CreateRepoResponse {
	s.Headers = v
	return s
}

type CreateRepoBuildRuleResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CreateRepoBuildRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateRepoBuildRuleResponse) GoString() string {
	return s.String()
}

func (s *CreateRepoBuildRuleResponse) SetHeaders(v map[string]*string) *CreateRepoBuildRuleResponse {
	s.Headers = v
	return s
}

type CreateRepoWebhookResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CreateRepoWebhookResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateRepoWebhookResponse) GoString() string {
	return s.String()
}

func (s *CreateRepoWebhookResponse) SetHeaders(v map[string]*string) *CreateRepoWebhookResponse {
	s.Headers = v
	return s
}

type CreateUserInfoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s CreateUserInfoResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateUserInfoResponse) GoString() string {
	return s.String()
}

func (s *CreateUserInfoResponse) SetHeaders(v map[string]*string) *CreateUserInfoResponse {
	s.Headers = v
	return s
}

type DeleteImageResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s DeleteImageResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteImageResponse) GoString() string {
	return s.String()
}

func (s *DeleteImageResponse) SetHeaders(v map[string]*string) *DeleteImageResponse {
	s.Headers = v
	return s
}

type DeleteNamespaceResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s DeleteNamespaceResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteNamespaceResponse) GoString() string {
	return s.String()
}

func (s *DeleteNamespaceResponse) SetHeaders(v map[string]*string) *DeleteNamespaceResponse {
	s.Headers = v
	return s
}

type DeleteRepoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s DeleteRepoResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteRepoResponse) GoString() string {
	return s.String()
}

func (s *DeleteRepoResponse) SetHeaders(v map[string]*string) *DeleteRepoResponse {
	s.Headers = v
	return s
}

type DeleteRepoBuildRuleResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s DeleteRepoBuildRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteRepoBuildRuleResponse) GoString() string {
	return s.String()
}

func (s *DeleteRepoBuildRuleResponse) SetHeaders(v map[string]*string) *DeleteRepoBuildRuleResponse {
	s.Headers = v
	return s
}

type DeleteRepoWebhookResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s DeleteRepoWebhookResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteRepoWebhookResponse) GoString() string {
	return s.String()
}

func (s *DeleteRepoWebhookResponse) SetHeaders(v map[string]*string) *DeleteRepoWebhookResponse {
	s.Headers = v
	return s
}

type GetAuthorizationTokenResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetAuthorizationTokenResponse) String() string {
	return tea.Prettify(s)
}

func (s GetAuthorizationTokenResponse) GoString() string {
	return s.String()
}

func (s *GetAuthorizationTokenResponse) SetHeaders(v map[string]*string) *GetAuthorizationTokenResponse {
	s.Headers = v
	return s
}

type GetImageLayerResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetImageLayerResponse) String() string {
	return tea.Prettify(s)
}

func (s GetImageLayerResponse) GoString() string {
	return s.String()
}

func (s *GetImageLayerResponse) SetHeaders(v map[string]*string) *GetImageLayerResponse {
	s.Headers = v
	return s
}

type GetImageManifestRequest struct {
	SchemaVersion *int32 `json:"SchemaVersion,omitempty" xml:"SchemaVersion,omitempty"`
}

func (s GetImageManifestRequest) String() string {
	return tea.Prettify(s)
}

func (s GetImageManifestRequest) GoString() string {
	return s.String()
}

func (s *GetImageManifestRequest) SetSchemaVersion(v int32) *GetImageManifestRequest {
	s.SchemaVersion = &v
	return s
}

type GetImageManifestResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetImageManifestResponse) String() string {
	return tea.Prettify(s)
}

func (s GetImageManifestResponse) GoString() string {
	return s.String()
}

func (s *GetImageManifestResponse) SetHeaders(v map[string]*string) *GetImageManifestResponse {
	s.Headers = v
	return s
}

type GetNamespaceResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetNamespaceResponse) String() string {
	return tea.Prettify(s)
}

func (s GetNamespaceResponse) GoString() string {
	return s.String()
}

func (s *GetNamespaceResponse) SetHeaders(v map[string]*string) *GetNamespaceResponse {
	s.Headers = v
	return s
}

type GetNamespaceListRequest struct {
	Authorize *string `json:"Authorize,omitempty" xml:"Authorize,omitempty"`
	Status    *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s GetNamespaceListRequest) String() string {
	return tea.Prettify(s)
}

func (s GetNamespaceListRequest) GoString() string {
	return s.String()
}

func (s *GetNamespaceListRequest) SetAuthorize(v string) *GetNamespaceListRequest {
	s.Authorize = &v
	return s
}

func (s *GetNamespaceListRequest) SetStatus(v string) *GetNamespaceListRequest {
	s.Status = &v
	return s
}

type GetNamespaceListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetNamespaceListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetNamespaceListResponse) GoString() string {
	return s.String()
}

func (s *GetNamespaceListResponse) SetHeaders(v map[string]*string) *GetNamespaceListResponse {
	s.Headers = v
	return s
}

type GetRegionRequest struct {
	Domain *string `json:"Domain,omitempty" xml:"Domain,omitempty"`
}

func (s GetRegionRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRegionRequest) GoString() string {
	return s.String()
}

func (s *GetRegionRequest) SetDomain(v string) *GetRegionRequest {
	s.Domain = &v
	return s
}

type GetRegionResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRegionResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRegionResponse) GoString() string {
	return s.String()
}

func (s *GetRegionResponse) SetHeaders(v map[string]*string) *GetRegionResponse {
	s.Headers = v
	return s
}

type GetRegionListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRegionListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRegionListResponse) GoString() string {
	return s.String()
}

func (s *GetRegionListResponse) SetHeaders(v map[string]*string) *GetRegionListResponse {
	s.Headers = v
	return s
}

type GetRepoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoResponse) GoString() string {
	return s.String()
}

func (s *GetRepoResponse) SetHeaders(v map[string]*string) *GetRepoResponse {
	s.Headers = v
	return s
}

type GetRepoBuildListRequest struct {
	Page     *int32 `json:"Page,omitempty" xml:"Page,omitempty"`
	PageSize *int32 `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
}

func (s GetRepoBuildListRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRepoBuildListRequest) GoString() string {
	return s.String()
}

func (s *GetRepoBuildListRequest) SetPage(v int32) *GetRepoBuildListRequest {
	s.Page = &v
	return s
}

func (s *GetRepoBuildListRequest) SetPageSize(v int32) *GetRepoBuildListRequest {
	s.PageSize = &v
	return s
}

type GetRepoBuildListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoBuildListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoBuildListResponse) GoString() string {
	return s.String()
}

func (s *GetRepoBuildListResponse) SetHeaders(v map[string]*string) *GetRepoBuildListResponse {
	s.Headers = v
	return s
}

type GetRepoBuildRuleListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoBuildRuleListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoBuildRuleListResponse) GoString() string {
	return s.String()
}

func (s *GetRepoBuildRuleListResponse) SetHeaders(v map[string]*string) *GetRepoBuildRuleListResponse {
	s.Headers = v
	return s
}

type GetRepoBuildStatusResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoBuildStatusResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoBuildStatusResponse) GoString() string {
	return s.String()
}

func (s *GetRepoBuildStatusResponse) SetHeaders(v map[string]*string) *GetRepoBuildStatusResponse {
	s.Headers = v
	return s
}

type GetRepoListRequest struct {
	Page     *int32  `json:"Page,omitempty" xml:"Page,omitempty"`
	PageSize *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	Status   *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s GetRepoListRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRepoListRequest) GoString() string {
	return s.String()
}

func (s *GetRepoListRequest) SetPage(v int32) *GetRepoListRequest {
	s.Page = &v
	return s
}

func (s *GetRepoListRequest) SetPageSize(v int32) *GetRepoListRequest {
	s.PageSize = &v
	return s
}

func (s *GetRepoListRequest) SetStatus(v string) *GetRepoListRequest {
	s.Status = &v
	return s
}

type GetRepoListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoListResponse) GoString() string {
	return s.String()
}

func (s *GetRepoListResponse) SetHeaders(v map[string]*string) *GetRepoListResponse {
	s.Headers = v
	return s
}

type GetRepoListByNamespaceRequest struct {
	Page     *int32  `json:"Page,omitempty" xml:"Page,omitempty"`
	PageSize *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	Status   *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s GetRepoListByNamespaceRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRepoListByNamespaceRequest) GoString() string {
	return s.String()
}

func (s *GetRepoListByNamespaceRequest) SetPage(v int32) *GetRepoListByNamespaceRequest {
	s.Page = &v
	return s
}

func (s *GetRepoListByNamespaceRequest) SetPageSize(v int32) *GetRepoListByNamespaceRequest {
	s.PageSize = &v
	return s
}

func (s *GetRepoListByNamespaceRequest) SetStatus(v string) *GetRepoListByNamespaceRequest {
	s.Status = &v
	return s
}

type GetRepoListByNamespaceResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoListByNamespaceResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoListByNamespaceResponse) GoString() string {
	return s.String()
}

func (s *GetRepoListByNamespaceResponse) SetHeaders(v map[string]*string) *GetRepoListByNamespaceResponse {
	s.Headers = v
	return s
}

type GetRepoTagResponseBody struct {
	Digest      *string `json:"digest,omitempty" xml:"digest,omitempty"`
	ImageCreate *int64  `json:"imageCreate,omitempty" xml:"imageCreate,omitempty"`
	ImageId     *string `json:"imageId,omitempty" xml:"imageId,omitempty"`
	ImageSize   *int64  `json:"imageSize,omitempty" xml:"imageSize,omitempty"`
	ImageUpdate *int64  `json:"imageUpdate,omitempty" xml:"imageUpdate,omitempty"`
	RequestId   *string `json:"requestId,omitempty" xml:"requestId,omitempty"`
	Status      *string `json:"status,omitempty" xml:"status,omitempty"`
	Tag         *string `json:"tag,omitempty" xml:"tag,omitempty"`
}

func (s GetRepoTagResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagResponseBody) GoString() string {
	return s.String()
}

func (s *GetRepoTagResponseBody) SetDigest(v string) *GetRepoTagResponseBody {
	s.Digest = &v
	return s
}

func (s *GetRepoTagResponseBody) SetImageCreate(v int64) *GetRepoTagResponseBody {
	s.ImageCreate = &v
	return s
}

func (s *GetRepoTagResponseBody) SetImageId(v string) *GetRepoTagResponseBody {
	s.ImageId = &v
	return s
}

func (s *GetRepoTagResponseBody) SetImageSize(v int64) *GetRepoTagResponseBody {
	s.ImageSize = &v
	return s
}

func (s *GetRepoTagResponseBody) SetImageUpdate(v int64) *GetRepoTagResponseBody {
	s.ImageUpdate = &v
	return s
}

func (s *GetRepoTagResponseBody) SetRequestId(v string) *GetRepoTagResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetRepoTagResponseBody) SetStatus(v string) *GetRepoTagResponseBody {
	s.Status = &v
	return s
}

func (s *GetRepoTagResponseBody) SetTag(v string) *GetRepoTagResponseBody {
	s.Tag = &v
	return s
}

type GetRepoTagResponse struct {
	Headers map[string]*string      `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	Body    *GetRepoTagResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetRepoTagResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagResponse) GoString() string {
	return s.String()
}

func (s *GetRepoTagResponse) SetHeaders(v map[string]*string) *GetRepoTagResponse {
	s.Headers = v
	return s
}

func (s *GetRepoTagResponse) SetBody(v *GetRepoTagResponseBody) *GetRepoTagResponse {
	s.Body = v
	return s
}

type GetRepoTagScanListRequest struct {
	Page     *int32  `json:"Page,omitempty" xml:"Page,omitempty"`
	PageSize *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	Severity *string `json:"Severity,omitempty" xml:"Severity,omitempty"`
}

func (s GetRepoTagScanListRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagScanListRequest) GoString() string {
	return s.String()
}

func (s *GetRepoTagScanListRequest) SetPage(v int32) *GetRepoTagScanListRequest {
	s.Page = &v
	return s
}

func (s *GetRepoTagScanListRequest) SetPageSize(v int32) *GetRepoTagScanListRequest {
	s.PageSize = &v
	return s
}

func (s *GetRepoTagScanListRequest) SetSeverity(v string) *GetRepoTagScanListRequest {
	s.Severity = &v
	return s
}

type GetRepoTagScanListResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoTagScanListResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagScanListResponse) GoString() string {
	return s.String()
}

func (s *GetRepoTagScanListResponse) SetHeaders(v map[string]*string) *GetRepoTagScanListResponse {
	s.Headers = v
	return s
}

type GetRepoTagScanStatusResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoTagScanStatusResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagScanStatusResponse) GoString() string {
	return s.String()
}

func (s *GetRepoTagScanStatusResponse) SetHeaders(v map[string]*string) *GetRepoTagScanStatusResponse {
	s.Headers = v
	return s
}

type GetRepoTagScanSummaryResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoTagScanSummaryResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagScanSummaryResponse) GoString() string {
	return s.String()
}

func (s *GetRepoTagScanSummaryResponse) SetHeaders(v map[string]*string) *GetRepoTagScanSummaryResponse {
	s.Headers = v
	return s
}

type GetRepoTagsRequest struct {
	Page     *int32 `json:"Page,omitempty" xml:"Page,omitempty"`
	PageSize *int32 `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
}

func (s GetRepoTagsRequest) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagsRequest) GoString() string {
	return s.String()
}

func (s *GetRepoTagsRequest) SetPage(v int32) *GetRepoTagsRequest {
	s.Page = &v
	return s
}

func (s *GetRepoTagsRequest) SetPageSize(v int32) *GetRepoTagsRequest {
	s.PageSize = &v
	return s
}

type GetRepoTagsResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoTagsResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoTagsResponse) GoString() string {
	return s.String()
}

func (s *GetRepoTagsResponse) SetHeaders(v map[string]*string) *GetRepoTagsResponse {
	s.Headers = v
	return s
}

type GetRepoWebhookResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetRepoWebhookResponse) String() string {
	return tea.Prettify(s)
}

func (s GetRepoWebhookResponse) GoString() string {
	return s.String()
}

func (s *GetRepoWebhookResponse) SetHeaders(v map[string]*string) *GetRepoWebhookResponse {
	s.Headers = v
	return s
}

type GetResourceQuotaResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s GetResourceQuotaResponse) String() string {
	return tea.Prettify(s)
}

func (s GetResourceQuotaResponse) GoString() string {
	return s.String()
}

func (s *GetResourceQuotaResponse) SetHeaders(v map[string]*string) *GetResourceQuotaResponse {
	s.Headers = v
	return s
}

type StartImageScanResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s StartImageScanResponse) String() string {
	return tea.Prettify(s)
}

func (s StartImageScanResponse) GoString() string {
	return s.String()
}

func (s *StartImageScanResponse) SetHeaders(v map[string]*string) *StartImageScanResponse {
	s.Headers = v
	return s
}

type StartRepoBuildByRuleResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s StartRepoBuildByRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s StartRepoBuildByRuleResponse) GoString() string {
	return s.String()
}

func (s *StartRepoBuildByRuleResponse) SetHeaders(v map[string]*string) *StartRepoBuildByRuleResponse {
	s.Headers = v
	return s
}

type UpdateNamespaceResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s UpdateNamespaceResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateNamespaceResponse) GoString() string {
	return s.String()
}

func (s *UpdateNamespaceResponse) SetHeaders(v map[string]*string) *UpdateNamespaceResponse {
	s.Headers = v
	return s
}

type UpdateRepoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s UpdateRepoResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateRepoResponse) GoString() string {
	return s.String()
}

func (s *UpdateRepoResponse) SetHeaders(v map[string]*string) *UpdateRepoResponse {
	s.Headers = v
	return s
}

type UpdateRepoBuildRuleResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s UpdateRepoBuildRuleResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateRepoBuildRuleResponse) GoString() string {
	return s.String()
}

func (s *UpdateRepoBuildRuleResponse) SetHeaders(v map[string]*string) *UpdateRepoBuildRuleResponse {
	s.Headers = v
	return s
}

type UpdateRepoWebhookResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s UpdateRepoWebhookResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateRepoWebhookResponse) GoString() string {
	return s.String()
}

func (s *UpdateRepoWebhookResponse) SetHeaders(v map[string]*string) *UpdateRepoWebhookResponse {
	s.Headers = v
	return s
}

type UpdateUserInfoResponse struct {
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
}

func (s UpdateUserInfoResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateUserInfoResponse) GoString() string {
	return s.String()
}

func (s *UpdateUserInfoResponse) SetHeaders(v map[string]*string) *UpdateUserInfoResponse {
	s.Headers = v
	return s
}

type Client struct {
	openapi.Client
}

func NewClient(config *openapi.Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *openapi.Config) (_err error) {
	_err = client.Client.Init(config)
	if _err != nil {
		return _err
	}
	client.EndpointRule = tea.String("regional")
	_err = client.CheckConfig(config)
	if _err != nil {
		return _err
	}
	client.Endpoint, _err = client.GetEndpoint(tea.String("cr"), client.RegionId, client.EndpointRule, client.Network, client.Suffix, client.EndpointMap, client.Endpoint)
	if _err != nil {
		return _err
	}

	return nil
}

func (client *Client) GetEndpoint(productId *string, regionId *string, endpointRule *string, network *string, suffix *string, endpointMap map[string]*string, endpoint *string) (_result *string, _err error) {
	if !tea.BoolValue(util.Empty(endpoint)) {
		_result = endpoint
		return _result, _err
	}

	if !tea.BoolValue(util.IsUnset(endpointMap)) && !tea.BoolValue(util.Empty(endpointMap[tea.StringValue(regionId)])) {
		_result = endpointMap[tea.StringValue(regionId)]
		return _result, _err
	}

	_body, _err := endpointutil.GetEndpointRules(productId, regionId, endpointRule, network, suffix)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CancelRepoBuild(RepoNamespace *string, RepoName *string, BuildId *string) (_result *CancelRepoBuildResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CancelRepoBuildResponse{}
	_body, _err := client.CancelRepoBuildWithOptions(RepoNamespace, RepoName, BuildId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CancelRepoBuildWithOptions(RepoNamespace *string, RepoName *string, BuildId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CancelRepoBuildResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	BuildId = openapiutil.GetEncodeParam(BuildId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CancelRepoBuild"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/build/" + tea.StringValue(BuildId) + "/cancel"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CancelRepoBuildResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateNamespace() (_result *CreateNamespaceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateNamespaceResponse{}
	_body, _err := client.CreateNamespaceWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateNamespaceWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateNamespaceResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CreateNamespace"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/namespace"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CreateNamespaceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateRepo() (_result *CreateRepoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateRepoResponse{}
	_body, _err := client.CreateRepoWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateRepoWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateRepoResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CreateRepo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CreateRepoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateRepoBuildRule(RepoNamespace *string, RepoName *string) (_result *CreateRepoBuildRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateRepoBuildRuleResponse{}
	_body, _err := client.CreateRepoBuildRuleWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateRepoBuildRuleWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateRepoBuildRuleResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CreateRepoBuildRule"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/rules"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CreateRepoBuildRuleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateRepoWebhook(RepoNamespace *string, RepoName *string) (_result *CreateRepoWebhookResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateRepoWebhookResponse{}
	_body, _err := client.CreateRepoWebhookWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateRepoWebhookWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateRepoWebhookResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CreateRepoWebhook"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/webhooks"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CreateRepoWebhookResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateUserInfo() (_result *CreateUserInfoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateUserInfoResponse{}
	_body, _err := client.CreateUserInfoWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateUserInfoWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateUserInfoResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("CreateUserInfo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/users"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &CreateUserInfoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteImage(RepoNamespace *string, RepoName *string, Tag *string) (_result *DeleteImageResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteImageResponse{}
	_body, _err := client.DeleteImageWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteImageWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteImageResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteImage"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag)),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &DeleteImageResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteNamespace(Namespace *string) (_result *DeleteNamespaceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteNamespaceResponse{}
	_body, _err := client.DeleteNamespaceWithOptions(Namespace, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteNamespaceWithOptions(Namespace *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteNamespaceResponse, _err error) {
	Namespace = openapiutil.GetEncodeParam(Namespace)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteNamespace"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/namespace/" + tea.StringValue(Namespace)),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &DeleteNamespaceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteRepo(RepoNamespace *string, RepoName *string) (_result *DeleteRepoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteRepoResponse{}
	_body, _err := client.DeleteRepoWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteRepoWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteRepoResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteRepo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName)),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &DeleteRepoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteRepoBuildRule(RepoNamespace *string, RepoName *string, BuildRuleId *string) (_result *DeleteRepoBuildRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteRepoBuildRuleResponse{}
	_body, _err := client.DeleteRepoBuildRuleWithOptions(RepoNamespace, RepoName, BuildRuleId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteRepoBuildRuleWithOptions(RepoNamespace *string, RepoName *string, BuildRuleId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteRepoBuildRuleResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	BuildRuleId = openapiutil.GetEncodeParam(BuildRuleId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteRepoBuildRule"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/rules/" + tea.StringValue(BuildRuleId)),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &DeleteRepoBuildRuleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteRepoWebhook(RepoNamespace *string, RepoName *string, WebhookId *string) (_result *DeleteRepoWebhookResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteRepoWebhookResponse{}
	_body, _err := client.DeleteRepoWebhookWithOptions(RepoNamespace, RepoName, WebhookId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteRepoWebhookWithOptions(RepoNamespace *string, RepoName *string, WebhookId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteRepoWebhookResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	WebhookId = openapiutil.GetEncodeParam(WebhookId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteRepoWebhook"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/webhooks/" + tea.StringValue(WebhookId)),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &DeleteRepoWebhookResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetAuthorizationToken() (_result *GetAuthorizationTokenResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetAuthorizationTokenResponse{}
	_body, _err := client.GetAuthorizationTokenWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetAuthorizationTokenWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetAuthorizationTokenResponse, _err error) {
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
		BodyType:    tea.String("none"),
	}
	_result = &GetAuthorizationTokenResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetImageLayer(RepoNamespace *string, RepoName *string, Tag *string) (_result *GetImageLayerResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetImageLayerResponse{}
	_body, _err := client.GetImageLayerWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetImageLayerWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetImageLayerResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetImageLayer"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/layers"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetImageLayerResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetImageManifest(RepoNamespace *string, RepoName *string, Tag *string, request *GetImageManifestRequest) (_result *GetImageManifestResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetImageManifestResponse{}
	_body, _err := client.GetImageManifestWithOptions(RepoNamespace, RepoName, Tag, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetImageManifestWithOptions(RepoNamespace *string, RepoName *string, Tag *string, request *GetImageManifestRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetImageManifestResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.SchemaVersion)) {
		query["SchemaVersion"] = request.SchemaVersion
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetImageManifest"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/manifest"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetImageManifestResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetNamespace(Namespace *string) (_result *GetNamespaceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetNamespaceResponse{}
	_body, _err := client.GetNamespaceWithOptions(Namespace, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetNamespaceWithOptions(Namespace *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetNamespaceResponse, _err error) {
	Namespace = openapiutil.GetEncodeParam(Namespace)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetNamespace"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/namespace/" + tea.StringValue(Namespace)),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetNamespaceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetNamespaceList(request *GetNamespaceListRequest) (_result *GetNamespaceListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetNamespaceListResponse{}
	_body, _err := client.GetNamespaceListWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetNamespaceListWithOptions(request *GetNamespaceListRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetNamespaceListResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Authorize)) {
		query["Authorize"] = request.Authorize
	}

	if !tea.BoolValue(util.IsUnset(request.Status)) {
		query["Status"] = request.Status
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetNamespaceList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/namespace"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetNamespaceListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRegion(request *GetRegionRequest) (_result *GetRegionResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRegionResponse{}
	_body, _err := client.GetRegionWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRegionWithOptions(request *GetRegionRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRegionResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Domain)) {
		query["Domain"] = request.Domain
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRegion"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/regions"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRegionResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRegionList() (_result *GetRegionListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRegionListResponse{}
	_body, _err := client.GetRegionListWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRegionListWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRegionListResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRegionList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/regions"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRegionListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepo(RepoNamespace *string, RepoName *string) (_result *GetRepoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoResponse{}
	_body, _err := client.GetRepoWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName)),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoBuildList(RepoNamespace *string, RepoName *string, request *GetRepoBuildListRequest) (_result *GetRepoBuildListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoBuildListResponse{}
	_body, _err := client.GetRepoBuildListWithOptions(RepoNamespace, RepoName, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoBuildListWithOptions(RepoNamespace *string, RepoName *string, request *GetRepoBuildListRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoBuildListResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Page)) {
		query["Page"] = request.Page
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoBuildList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/build"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoBuildListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoBuildRuleList(RepoNamespace *string, RepoName *string) (_result *GetRepoBuildRuleListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoBuildRuleListResponse{}
	_body, _err := client.GetRepoBuildRuleListWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoBuildRuleListWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoBuildRuleListResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoBuildRuleList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/rules"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoBuildRuleListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoBuildStatus(RepoNamespace *string, RepoName *string, BuildId *string) (_result *GetRepoBuildStatusResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoBuildStatusResponse{}
	_body, _err := client.GetRepoBuildStatusWithOptions(RepoNamespace, RepoName, BuildId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoBuildStatusWithOptions(RepoNamespace *string, RepoName *string, BuildId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoBuildStatusResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	BuildId = openapiutil.GetEncodeParam(BuildId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoBuildStatus"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/build/" + tea.StringValue(BuildId) + "/status"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoBuildStatusResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoList(request *GetRepoListRequest) (_result *GetRepoListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoListResponse{}
	_body, _err := client.GetRepoListWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoListWithOptions(request *GetRepoListRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoListResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Page)) {
		query["Page"] = request.Page
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.Status)) {
		query["Status"] = request.Status
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoListByNamespace(RepoNamespace *string, request *GetRepoListByNamespaceRequest) (_result *GetRepoListByNamespaceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoListByNamespaceResponse{}
	_body, _err := client.GetRepoListByNamespaceWithOptions(RepoNamespace, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoListByNamespaceWithOptions(RepoNamespace *string, request *GetRepoListByNamespaceRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoListByNamespaceResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Page)) {
		query["Page"] = request.Page
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.Status)) {
		query["Status"] = request.Status
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoListByNamespace"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace)),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoListByNamespaceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoTag(RepoNamespace *string, RepoName *string, Tag *string) (_result *GetRepoTagResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoTagResponse{}
	_body, _err := client.GetRepoTagWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoTagWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoTagResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoTag"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag)),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetRepoTagResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoTagScanList(RepoNamespace *string, RepoName *string, Tag *string, request *GetRepoTagScanListRequest) (_result *GetRepoTagScanListResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoTagScanListResponse{}
	_body, _err := client.GetRepoTagScanListWithOptions(RepoNamespace, RepoName, Tag, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoTagScanListWithOptions(RepoNamespace *string, RepoName *string, Tag *string, request *GetRepoTagScanListRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoTagScanListResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Page)) {
		query["Page"] = request.Page
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.Severity)) {
		query["Severity"] = request.Severity
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoTagScanList"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/scanResult"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoTagScanListResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoTagScanStatus(RepoNamespace *string, RepoName *string, Tag *string) (_result *GetRepoTagScanStatusResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoTagScanStatusResponse{}
	_body, _err := client.GetRepoTagScanStatusWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoTagScanStatusWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoTagScanStatusResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoTagScanStatus"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/scanStatus"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoTagScanStatusResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoTagScanSummary(RepoNamespace *string, RepoName *string, Tag *string) (_result *GetRepoTagScanSummaryResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoTagScanSummaryResponse{}
	_body, _err := client.GetRepoTagScanSummaryWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoTagScanSummaryWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoTagScanSummaryResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoTagScanSummary"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/scanCount"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoTagScanSummaryResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoTags(RepoNamespace *string, RepoName *string, request *GetRepoTagsRequest) (_result *GetRepoTagsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoTagsResponse{}
	_body, _err := client.GetRepoTagsWithOptions(RepoNamespace, RepoName, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoTagsWithOptions(RepoNamespace *string, RepoName *string, request *GetRepoTagsRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoTagsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Page)) {
		query["Page"] = request.Page
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoTags"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoTagsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetRepoWebhook(RepoNamespace *string, RepoName *string) (_result *GetRepoWebhookResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetRepoWebhookResponse{}
	_body, _err := client.GetRepoWebhookWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetRepoWebhookWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetRepoWebhookResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetRepoWebhook"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/webhooks"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetRepoWebhookResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetResourceQuota(ResourceName *string) (_result *GetResourceQuotaResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetResourceQuotaResponse{}
	_body, _err := client.GetResourceQuotaWithOptions(ResourceName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetResourceQuotaWithOptions(ResourceName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetResourceQuotaResponse, _err error) {
	ResourceName = openapiutil.GetEncodeParam(ResourceName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetResourceQuota"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/resource/" + tea.StringValue(ResourceName)),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &GetResourceQuotaResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) StartImageScan(RepoNamespace *string, RepoName *string, Tag *string) (_result *StartImageScanResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &StartImageScanResponse{}
	_body, _err := client.StartImageScanWithOptions(RepoNamespace, RepoName, Tag, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) StartImageScanWithOptions(RepoNamespace *string, RepoName *string, Tag *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *StartImageScanResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	Tag = openapiutil.GetEncodeParam(Tag)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("StartImageScan"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/tags/" + tea.StringValue(Tag) + "/scan"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &StartImageScanResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) StartRepoBuildByRule(RepoNamespace *string, RepoName *string, BuildRuleId *string) (_result *StartRepoBuildByRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &StartRepoBuildByRuleResponse{}
	_body, _err := client.StartRepoBuildByRuleWithOptions(RepoNamespace, RepoName, BuildRuleId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) StartRepoBuildByRuleWithOptions(RepoNamespace *string, RepoName *string, BuildRuleId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *StartRepoBuildByRuleResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	BuildRuleId = openapiutil.GetEncodeParam(BuildRuleId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("StartRepoBuildByRule"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/rules/" + tea.StringValue(BuildRuleId) + "/build"),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &StartRepoBuildByRuleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateNamespace(Namespace *string) (_result *UpdateNamespaceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateNamespaceResponse{}
	_body, _err := client.UpdateNamespaceWithOptions(Namespace, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateNamespaceWithOptions(Namespace *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateNamespaceResponse, _err error) {
	Namespace = openapiutil.GetEncodeParam(Namespace)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateNamespace"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/namespace/" + tea.StringValue(Namespace)),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &UpdateNamespaceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateRepo(RepoNamespace *string, RepoName *string) (_result *UpdateRepoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateRepoResponse{}
	_body, _err := client.UpdateRepoWithOptions(RepoNamespace, RepoName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateRepoWithOptions(RepoNamespace *string, RepoName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateRepoResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateRepo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName)),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &UpdateRepoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateRepoBuildRule(RepoNamespace *string, RepoName *string, BuildRuleId *string) (_result *UpdateRepoBuildRuleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateRepoBuildRuleResponse{}
	_body, _err := client.UpdateRepoBuildRuleWithOptions(RepoNamespace, RepoName, BuildRuleId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateRepoBuildRuleWithOptions(RepoNamespace *string, RepoName *string, BuildRuleId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateRepoBuildRuleResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	BuildRuleId = openapiutil.GetEncodeParam(BuildRuleId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateRepoBuildRule"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/rules/" + tea.StringValue(BuildRuleId)),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &UpdateRepoBuildRuleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateRepoWebhook(RepoNamespace *string, RepoName *string, WebhookId *string) (_result *UpdateRepoWebhookResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateRepoWebhookResponse{}
	_body, _err := client.UpdateRepoWebhookWithOptions(RepoNamespace, RepoName, WebhookId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateRepoWebhookWithOptions(RepoNamespace *string, RepoName *string, WebhookId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateRepoWebhookResponse, _err error) {
	RepoNamespace = openapiutil.GetEncodeParam(RepoNamespace)
	RepoName = openapiutil.GetEncodeParam(RepoName)
	WebhookId = openapiutil.GetEncodeParam(WebhookId)
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateRepoWebhook"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/repos/" + tea.StringValue(RepoNamespace) + "/" + tea.StringValue(RepoName) + "/webhooks/" + tea.StringValue(WebhookId)),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &UpdateRepoWebhookResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateUserInfo() (_result *UpdateUserInfoResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateUserInfoResponse{}
	_body, _err := client.UpdateUserInfoWithOptions(headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateUserInfoWithOptions(headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateUserInfoResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateUserInfo"),
		Version:     tea.String("2016-06-07"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/users"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("none"),
	}
	_result = &UpdateUserInfoResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}
