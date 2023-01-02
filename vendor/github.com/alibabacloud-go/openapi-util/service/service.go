// This file is auto-generated, don't edit it. Thanks.
/**
 * This is for OpenApi Util
 */
package service

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/tjfoc/gmsm/sm3"
)

const (
	PEM_BEGIN = "-----BEGIN RSA PRIVATE KEY-----\n"
	PEM_END   = "\n-----END RSA PRIVATE KEY-----"
)

type Sorter struct {
	Keys []string
	Vals []string
}

func newSorter(m map[string]string) *Sorter {
	hs := &Sorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

// Sort is an additional function for function SignHeader.
func (hs *Sorter) Sort() {
	sort.Sort(hs)
}

// Len is an additional function for function SignHeader.
func (hs *Sorter) Len() int {
	return len(hs.Vals)
}

// Less is an additional function for function SignHeader.
func (hs *Sorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

// Swap is an additional function for function SignHeader.
func (hs *Sorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}

/**
 * Convert all params of body other than type of readable into content
 * @param body source Model
 * @param content target Model
 * @return void
 */
func Convert(body interface{}, content interface{}) {
	res := make(map[string]interface{})
	val := reflect.ValueOf(body).Elem()
	dataType := val.Type()
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		name, _ := field.Tag.Lookup("json")
		name = strings.Split(name, ",omitempty")[0]
		_, ok := val.Field(i).Interface().(io.Reader)
		if !ok {
			res[name] = val.Field(i).Interface()
		}
	}
	byt, _ := json.Marshal(res)
	json.Unmarshal(byt, content)
}

/**
 * Get the string to be signed according to request
 * @param request  which contains signed messages
 * @return the signed string
 */
func GetStringToSign(request *tea.Request) (_result *string) {
	return tea.String(getStringToSign(request))
}

func getStringToSign(request *tea.Request) string {
	resource := tea.StringValue(request.Pathname)
	queryParams := request.Query
	// sort QueryParams by key
	var queryKeys []string
	for key := range queryParams {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	tmp := ""
	for i := 0; i < len(queryKeys); i++ {
		queryKey := queryKeys[i]
		v := tea.StringValue(queryParams[queryKey])
		if v != "" {
			tmp = tmp + "&" + queryKey + "=" + v
		} else {
			tmp = tmp + "&" + queryKey
		}
	}
	if tmp != "" {
		tmp = strings.TrimLeft(tmp, "&")
		resource = resource + "?" + tmp
	}
	return getSignedStr(request, resource)
}

func getSignedStr(req *tea.Request, canonicalizedResource string) string {
	temp := make(map[string]string)

	for k, v := range req.Headers {
		if strings.HasPrefix(strings.ToLower(k), "x-acs-") {
			temp[strings.ToLower(k)] = tea.StringValue(v)
		}
	}
	hs := newSorter(temp)

	// Sort the temp by the ascending order
	hs.Sort()

	// Get the canonicalizedOSSHeaders
	canonicalizedOSSHeaders := ""
	for i := range hs.Keys {
		canonicalizedOSSHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	// Give other parameters values
	// when sign URL, date is expires
	date := tea.StringValue(req.Headers["date"])
	accept := tea.StringValue(req.Headers["accept"])
	contentType := tea.StringValue(req.Headers["content-type"])
	contentMd5 := tea.StringValue(req.Headers["content-md5"])

	signStr := tea.StringValue(req.Method) + "\n" + accept + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedOSSHeaders + canonicalizedResource
	return signStr
}

/**
 * Get signature according to stringToSign, secret
 * @param stringToSign  the signed string
 * @param secret accesskey secret
 * @return the signature
 */
func GetROASignature(stringToSign *string, secret *string) (_result *string) {
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(tea.StringValue(secret)))
	io.WriteString(h, tea.StringValue(stringToSign))
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return tea.String(signedStr)
}

func GetEndpoint(endpoint *string, server *bool, endpointType *string) *string {
	if tea.StringValue(endpointType) == "internal" {
		strs := strings.Split(tea.StringValue(endpoint), ".")
		strs[0] += "-internal"
		endpoint = tea.String(strings.Join(strs, "."))
	}
	if tea.BoolValue(server) && tea.StringValue(endpointType) == "accelerate" {
		return tea.String("oss-accelerate.aliyuncs.com")
	}

	return endpoint
}

func HexEncode(raw []byte) *string {
	return tea.String(hex.EncodeToString(raw))
}

func Hash(raw []byte, signatureAlgorithm *string) []byte {
	signType := tea.StringValue(signatureAlgorithm)
	if signType == "ACS3-HMAC-SHA256" || signType == "ACS3-RSA-SHA256" {
		h := sha256.New()
		h.Write(raw)
		return h.Sum(nil)
	} else if signType == "ACS3-HMAC-SM3" {
		h := sm3.New()
		h.Write(raw)
		return h.Sum(nil)
	}
	return nil
}

func GetEncodePath(path *string) *string {
	uri := tea.StringValue(path)
	strs := strings.Split(uri, "/")
	for i, v := range strs {
		strs[i] = url.QueryEscape(v)
	}
	uri = strings.Join(strs, "/")
	uri = strings.Replace(uri, "+", "%20", -1)
	uri = strings.Replace(uri, "*", "%2A", -1)
	uri = strings.Replace(uri, "%7E", "~", -1)
	return tea.String(uri)
}

func GetEncodeParam(param *string) *string {
	uri := tea.StringValue(param)
	uri = url.QueryEscape(uri)
	uri = strings.Replace(uri, "+", "%20", -1)
	uri = strings.Replace(uri, "*", "%2A", -1)
	uri = strings.Replace(uri, "%7E", "~", -1)
	return tea.String(uri)
}

func GetAuthorization(request *tea.Request, signatureAlgorithm, payload, acesskey, secret *string) *string {
	canonicalURI := tea.StringValue(request.Pathname)
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalURI = strings.Replace(canonicalURI, "+", "%20", -1)
	canonicalURI = strings.Replace(canonicalURI, "*", "%2A", -1)
	canonicalURI = strings.Replace(canonicalURI, "%7E", "~", -1)

	method := tea.StringValue(request.Method)
	canonicalQueryString := getCanonicalQueryString(request.Query)
	canonicalheaders, signedHeaders := getCanonicalHeaders(request.Headers)

	canonicalRequest := method + "\n" + canonicalURI + "\n" + canonicalQueryString + "\n" + canonicalheaders + "\n" +
		strings.Join(signedHeaders, ";") + "\n" + tea.StringValue(payload)
	signType := tea.StringValue(signatureAlgorithm)
	StringToSign := signType + "\n" + tea.StringValue(HexEncode(Hash([]byte(canonicalRequest), signatureAlgorithm)))
	signature := tea.StringValue(HexEncode(SignatureMethod(tea.StringValue(secret), StringToSign, signType)))
	auth := signType + " Credential=" + tea.StringValue(acesskey) + ",SignedHeaders=" +
		strings.Join(signedHeaders, ";") + ",Signature=" + signature
	return tea.String(auth)
}

func SignatureMethod(secret, source, signatureAlgorithm string) []byte {
	if signatureAlgorithm == "ACS3-HMAC-SHA256" {
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(source))
		return h.Sum(nil)
	} else if signatureAlgorithm == "ACS3-HMAC-SM3" {
		h := hmac.New(sm3.New, []byte(secret))
		h.Write([]byte(source))
		return h.Sum(nil)
	} else if signatureAlgorithm == "ACS3-RSA-SHA256" {
		return rsaSign(source, secret)
	}
	return nil
}

func rsaSign(content, secret string) []byte {
	h := crypto.SHA256.New()
	h.Write([]byte(content))
	hashed := h.Sum(nil)
	priv, err := parsePrivateKey(secret)
	if err != nil {
		return nil
	}
	sign, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
	if err != nil {
		return nil
	}
	return sign
}

func parsePrivateKey(privateKey string) (*rsa.PrivateKey, error) {
	privateKey = formatPrivateKey(privateKey)
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("PrivateKey is invalid")
	}
	priKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch priKey.(type) {
	case *rsa.PrivateKey:
		return priKey.(*rsa.PrivateKey), nil
	default:
		return nil, nil
	}
}

func formatPrivateKey(privateKey string) string {
	if !strings.HasPrefix(privateKey, PEM_BEGIN) {
		privateKey = PEM_BEGIN + privateKey
	}

	if !strings.HasSuffix(privateKey, PEM_END) {
		privateKey += PEM_END
	}
	return privateKey
}

func getCanonicalHeaders(headers map[string]*string) (string, []string) {
	tmp := make(map[string]string)
	tmpHeader := http.Header{}
	for k, v := range headers {
		if strings.HasPrefix(strings.ToLower(k), "x-acs-") || strings.ToLower(k) == "host" ||
			strings.ToLower(k) == "content-type" {
			tmp[strings.ToLower(k)] = strings.TrimSpace(tea.StringValue(v))
			tmpHeader.Add(strings.ToLower(k), strings.TrimSpace(tea.StringValue(v)))
		}
	}
	hs := newSorter(tmp)

	// Sort the temp by the ascending order
	hs.Sort()
	canonicalheaders := ""
	for _, key := range hs.Keys {
		vals := tmpHeader[textproto.CanonicalMIMEHeaderKey(key)]
		sort.Strings(vals)
		canonicalheaders += key + ":" + strings.Join(vals, ",") + "\n"
	}

	return canonicalheaders, hs.Keys
}

func getCanonicalQueryString(query map[string]*string) string {
	canonicalQueryString := ""
	if tea.BoolValue(util.IsUnset(query)) {
		return canonicalQueryString
	}
	tmp := make(map[string]string)
	for k, v := range query {
		tmp[k] = tea.StringValue(v)
	}

	hs := newSorter(tmp)

	// Sort the temp by the ascending order
	hs.Sort()
	for i := range hs.Keys {
		if hs.Vals[i] != "" {
			canonicalQueryString += "&" + hs.Keys[i] + "=" + url.QueryEscape(hs.Vals[i])
		} else {
			canonicalQueryString += "&" + hs.Keys[i] + "="
		}
	}
	canonicalQueryString = strings.Replace(canonicalQueryString, "+", "%20", -1)
	canonicalQueryString = strings.Replace(canonicalQueryString, "*", "%2A", -1)
	canonicalQueryString = strings.Replace(canonicalQueryString, "%7E", "~", -1)

	if canonicalQueryString != "" {
		canonicalQueryString = strings.TrimLeft(canonicalQueryString, "&")
	}
	return canonicalQueryString
}

/**
 * Parse filter into a form string
 * @param filter object
 * @return the string
 */
func ToForm(filter map[string]interface{}) (_result *string) {
	tmp := make(map[string]interface{})
	byt, _ := json.Marshal(filter)
	d := json.NewDecoder(bytes.NewReader(byt))
	d.UseNumber()
	_ = d.Decode(&tmp)

	result := make(map[string]*string)
	for key, value := range tmp {
		filterValue := reflect.ValueOf(value)
		flatRepeatedList(filterValue, result, key)
	}

	m := util.AnyifyMapValue(result)
	return util.ToFormString(m)
}

func flatRepeatedList(dataValue reflect.Value, result map[string]*string, prefix string) {
	if !dataValue.IsValid() {
		return
	}

	dataType := dataValue.Type()
	if dataType.Kind().String() == "slice" {
		handleRepeatedParams(dataValue, result, prefix)
	} else if dataType.Kind().String() == "map" {
		handleMap(dataValue, result, prefix)
	} else {
		result[prefix] = tea.String(fmt.Sprintf("%v", dataValue.Interface()))
	}
}

func handleRepeatedParams(repeatedFieldValue reflect.Value, result map[string]*string, prefix string) {
	if repeatedFieldValue.IsValid() && !repeatedFieldValue.IsNil() {
		for m := 0; m < repeatedFieldValue.Len(); m++ {
			elementValue := repeatedFieldValue.Index(m)
			key := prefix + "." + strconv.Itoa(m+1)
			fieldValue := reflect.ValueOf(elementValue.Interface())
			if fieldValue.Kind().String() == "map" {
				handleMap(fieldValue, result, key)
			} else {
				result[key] = tea.String(fmt.Sprintf("%v", fieldValue.Interface()))
			}
		}
	}
}

func handleMap(valueField reflect.Value, result map[string]*string, prefix string) {
	if valueField.IsValid() && valueField.String() != "" {
		valueFieldType := valueField.Type()
		if valueFieldType.Kind().String() == "map" {
			var byt []byte
			byt, _ = json.Marshal(valueField.Interface())
			cache := make(map[string]interface{})
			d := json.NewDecoder(bytes.NewReader(byt))
			d.UseNumber()
			_ = d.Decode(&cache)
			for key, value := range cache {
				pre := ""
				if prefix != "" {
					pre = prefix + "." + key
				} else {
					pre = key
				}
				fieldValue := reflect.ValueOf(value)
				flatRepeatedList(fieldValue, result, pre)
			}
		}
	}
}

/**
 * Get timestamp
 * @return the timestamp string
 */
func GetTimestamp() (_result *string) {
	gmt := time.FixedZone("GMT", 0)
	return tea.String(time.Now().In(gmt).Format("2006-01-02T15:04:05Z"))
}

/**
 * Parse filter into a object which's type is map[string]string
 * @param filter query param
 * @return the object
 */
func Query(filter interface{}) (_result map[string]*string) {
	tmp := make(map[string]interface{})
	byt, _ := json.Marshal(filter)
	d := json.NewDecoder(bytes.NewReader(byt))
	d.UseNumber()
	_ = d.Decode(&tmp)

	result := make(map[string]*string)
	for key, value := range tmp {
		filterValue := reflect.ValueOf(value)
		flatRepeatedList(filterValue, result, key)
	}

	return result
}

/**
 * Get signature according to signedParams, method and secret
 * @param signedParams params which need to be signed
 * @param method http method e.g. GET
 * @param secret AccessKeySecret
 * @return the signature
 */
func GetRPCSignature(signedParams map[string]*string, method *string, secret *string) (_result *string) {
	stringToSign := buildRpcStringToSign(signedParams, tea.StringValue(method))
	signature := sign(stringToSign, tea.StringValue(secret), "&")
	return tea.String(signature)
}

/**
 * Parse array into a string with specified style
 * @param array the array
 * @param prefix the prefix string
 * @style specified style e.g. repeatList
 * @return the string
 */
func ArrayToStringWithSpecifiedStyle(array interface{}, prefix *string, style *string) (_result *string) {
	if tea.BoolValue(util.IsUnset(array)) {
		return tea.String("")
	}

	sty := tea.StringValue(style)
	if sty == "repeatList" {
		tmp := map[string]interface{}{
			tea.StringValue(prefix): array,
		}
		return flatRepeatList(tmp)
	} else if sty == "simple" || sty == "spaceDelimited" || sty == "pipeDelimited" {
		return flatArray(array, sty)
	} else if sty == "json" {
		return util.ToJSONString(array)
	}
	return tea.String("")
}

func ParseToMap(in interface{}) map[string]interface{} {
	if tea.BoolValue(util.IsUnset(in)) {
		return nil
	}

	tmp := make(map[string]interface{})
	byt, _ := json.Marshal(in)
	d := json.NewDecoder(bytes.NewReader(byt))
	d.UseNumber()
	err := d.Decode(&tmp)
	if err != nil {
		return nil
	}
	return tmp
}

func flatRepeatList(filter map[string]interface{}) (_result *string) {
	tmp := make(map[string]interface{})
	byt, _ := json.Marshal(filter)
	d := json.NewDecoder(bytes.NewReader(byt))
	d.UseNumber()
	_ = d.Decode(&tmp)

	result := make(map[string]*string)
	for key, value := range tmp {
		filterValue := reflect.ValueOf(value)
		flatRepeatedList(filterValue, result, key)
	}

	res := make(map[string]string)
	for k, v := range result {
		res[k] = tea.StringValue(v)
	}
	hs := newSorter(res)

	hs.Sort()

	// Get the canonicalizedOSSHeaders
	t := ""
	for i := range hs.Keys {
		if i == len(hs.Keys)-1 {
			t += hs.Keys[i] + "=" + hs.Vals[i]
		} else {
			t += hs.Keys[i] + "=" + hs.Vals[i] + "&&"
		}
	}
	return tea.String(t)
}

func flatArray(array interface{}, sty string) *string {
	t := reflect.ValueOf(array)
	strs := make([]string, 0)
	for i := 0; i < t.Len(); i++ {
		tmp := t.Index(i)
		if tmp.Kind() == reflect.Ptr || tmp.Kind() == reflect.Interface {
			tmp = tmp.Elem()
		}

		if tmp.Kind() == reflect.Ptr {
			tmp = tmp.Elem()
		}
		if tmp.Kind() == reflect.String {
			strs = append(strs, tmp.String())
		} else {
			inter := tmp.Interface()
			byt, _ := json.Marshal(inter)
			strs = append(strs, string(byt))
		}
	}
	str := ""
	if sty == "simple" {
		str = strings.Join(strs, ",")
	} else if sty == "spaceDelimited" {
		str = strings.Join(strs, " ")
	} else if sty == "pipeDelimited" {
		str = strings.Join(strs, "|")
	}
	return tea.String(str)
}

func buildRpcStringToSign(signedParam map[string]*string, method string) (stringToSign string) {
	signParams := make(map[string]string)
	for key, value := range signedParam {
		signParams[key] = tea.StringValue(value)
	}

	stringToSign = getUrlFormedMap(signParams)
	stringToSign = strings.Replace(stringToSign, "+", "%20", -1)
	stringToSign = strings.Replace(stringToSign, "*", "%2A", -1)
	stringToSign = strings.Replace(stringToSign, "%7E", "~", -1)
	stringToSign = url.QueryEscape(stringToSign)
	stringToSign = method + "&%2F&" + stringToSign
	return
}

func getUrlFormedMap(source map[string]string) (urlEncoded string) {
	urlEncoder := url.Values{}
	for key, value := range source {
		urlEncoder.Add(key, value)
	}
	urlEncoded = urlEncoder.Encode()
	return
}

func sign(stringToSign, accessKeySecret, secretSuffix string) string {
	secret := accessKeySecret + secretSuffix
	signedBytes := shaHmac1(stringToSign, secret)
	signedString := base64.StdEncoding.EncodeToString(signedBytes)
	return signedString
}

func shaHmac1(source, secret string) []byte {
	key := []byte(secret)
	hmac := hmac.New(sha1.New, key)
	hmac.Write([]byte(source))
	return hmac.Sum(nil)
}
