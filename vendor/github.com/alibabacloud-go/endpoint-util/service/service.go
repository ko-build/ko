// This file is auto-generated, don't edit it. Thanks.
/**
 * Get endpoint
 * @return string
 */
package service

import (
	"fmt"
	"strings"

	"github.com/alibabacloud-go/tea/tea"
)

func GetEndpointRules(product, regionId, endpointType, network, suffix *string) (_result *string, _err error) {
	if tea.StringValue(endpointType) == "regional" {
		if tea.StringValue(regionId) == "" {
			_err = fmt.Errorf("RegionId is empty, please set a valid RegionId")
			return tea.String(""), _err
		}
		_result = tea.String(strings.Replace("<product><suffix><network>.<region_id>.aliyuncs.com",
			"<region_id>", tea.StringValue(regionId), 1))
	} else {
		_result = tea.String("<product><suffix><network>.aliyuncs.com")
	}
	_result = tea.String(strings.Replace(tea.StringValue(_result),
		"<product>", strings.ToLower(tea.StringValue(product)), 1))
	if tea.StringValue(network) == "" || tea.StringValue(network) == "public" {
		_result = tea.String(strings.Replace(tea.StringValue(_result), "<network>", "", 1))
	} else {
		_result = tea.String(strings.Replace(tea.StringValue(_result),
			"<network>", "-"+tea.StringValue(network), 1))
	}
	if tea.StringValue(suffix) == "" {
		_result = tea.String(strings.Replace(tea.StringValue(_result), "<suffix>", "", 1))
	} else {
		_result = tea.String(strings.Replace(tea.StringValue(_result),
			"<suffix>", "-"+tea.StringValue(suffix), 1))
	}
	return _result, nil
}
