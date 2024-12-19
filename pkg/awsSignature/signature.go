package awsSignature

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"github.com/samber/lo"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// getCanonicalHeadersV4 generate a list of request headers with their values
func getCanonicalHeadersV4(signedHeaders http.Header) string {
	var headers []string
	vals := make(http.Header)
	for k, vv := range signedHeaders {
		headers = append(headers, strings.ToLower(k))
		vals[strings.ToLower(k)] = vv
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteByte(':')
		for idx, v := range vals[k] {
			if idx > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(signV4TrimAll(v))
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

// getSignedHeaders generate all signed request headers.
// i.e lexically sorted, semicolon-separated list of lowercase
// request header names.
func getSignedHeaders(req http.Request, ignoredHeaders map[string]bool) string {
	var headers []string
	for k := range req.Header {
		if _, ok := ignoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // Ignored header found continue.
		}
		headers = append(headers, strings.ToLower(k))
	}
	headers = append(headers, "host")
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// getSignedHeadersV4 generate a string i.e alphabetically sorted, semicolon-separated list of lowercase request header names
func getSignedHeadersV4(signedHeaders http.Header) string {
	var headers []string
	for k := range signedHeaders {
		headers = append(headers, strings.ToLower(k))
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// if object matches reserved string, no need to encode them
var reservedObjectNames = regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")

// EncodePath encode the strings from UTF-8 byte representations to HTML hex escape sequences
//
// This is necessary since regular url.Parse() and url.Encode() functions do not support UTF-8
// non english characters cannot be parsed due to the nature in which url.Encode() is written
//
// This function on the other hand is a direct replacement for url.Encode() technique to support
// pretty much every UTF-8 character.
func EncodePath(pathName string) string {
	if reservedObjectNames.MatchString(pathName) {
		return pathName
	}
	var encodedPathname strings.Builder
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedPathname.WriteRune(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/': // §2.3 Unreserved characters (mark)
			encodedPathname.WriteRune(s)
			continue
		default:
			_len := utf8.RuneLen(s)
			if _len < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, _len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				_hex := hex.EncodeToString([]byte{r})
				encodedPathname.WriteString("%" + strings.ToUpper(_hex))
			}
		}
	}
	return encodedPathname.String()
}

// getCanonicalRequestV4 generate a canonical request of style
//
// canonicalRequest =
//
//	<HTTPMethod>\n
//	<CanonicalURI>\n
//	<CanonicalQueryString>\n
//	<CanonicalHeaders>\n
//	<SignedHeaders>\n
//	<HashedPayload>
func getCanonicalRequestV4(extractedSignedHeaders http.Header, payload, queryStr, urlPath, method string) string {
	rawQuery := strings.Replace(queryStr, "+", "%20", -1)
	encodedPath := EncodePath(urlPath)
	canonicalRequest := strings.Join([]string{
		method,
		encodedPath,
		rawQuery,
		getCanonicalHeadersV4(extractedSignedHeaders),
		getSignedHeadersV4(extractedSignedHeaders),
		payload,
	}, "\n")
	return canonicalRequest
}

// getScopeV4 generate a string of a specific date, an AWS region, and a service.
func getScopeV4(t time.Time, region string) string {
	scope := strings.Join([]string{
		t.Format(yyyymmdd),
		region,
		string(ServiceS3),
		"aws4_request",
	}, SlashSeparator)
	return scope
}

// getStringToSignV4 a string based on selected query values.
func getStringToSignV4(canonicalRequest string, t time.Time, scope string) string {
	stringToSign := signV4Algorithm + "\n" + t.Format(iso8601Format) + "\n"
	stringToSign = stringToSign + scope + "\n"
	canonicalRequestBytes := sha256.Sum256([]byte(canonicalRequest))
	stringToSign = stringToSign + hex.EncodeToString(canonicalRequestBytes[:])
	return stringToSign
}

// getSigningKeyV4 hmac seed to calculate final signature.
func getSigningKeyV4(secretKey string, t time.Time, region string, sType string) []byte {
	date := sumHMAC([]byte("AWS4"+secretKey), []byte(t.Format(yyyymmdd)))
	regionBytes := sumHMAC(date, []byte(region))
	service := sumHMAC(regionBytes, []byte(sType))
	signingKey := sumHMAC(service, []byte("aws4_request"))
	return signingKey
}

// getSignature final signature in hexadecimal form.
func getSignature(signingKey []byte, stringToSign string) string {
	return hex.EncodeToString(sumHMAC(signingKey, []byte(stringToSign)))
}

// compareSignatureV4 returns true if and only if both signatures
// are equal. The signatures are expected to be HEX encoded strings
// according to the AWS S3 signature V4 spec.
func compareSignatureV4(sig1, sig2 string) bool {
	// The CTC using []byte(str) works because the hex encoding
	// is unique for a sequence of bytes. See also compareSignatureV2.
	return subtle.ConstantTimeCompare([]byte(sig1), []byte(sig2)) == 1
}

// VerifyPolicySignature - Verify query headers with post policy
//
//   - http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-HTTPPOSTConstructPolicy.html
//
//     X-Amz-Credential 主要用于特定场景和兼容性需求（比如：AWS STS(Security Token Service)生成临时凭证时）
//
// returns ErrNone if the signature matches.
func (s *signer) VerifyPolicySignature(formValues http.Header) error {
	// Parse credential tag.
	credHeader, s3Err := parseCredentialHeaderV4("Credential=" + formValues.Get(AmzCredential))
	if s3Err != nil {
		return s3Err
	} else if !s.v4.isValidRegion(credHeader.scope.region) { // Should validate region, only if region is set.
		return ErrInvalidRegion
	}

	// Get signing key.
	signingKey := getSigningKeyV4(s.secretKey, credHeader.scope.date, credHeader.scope.region, ServiceS3)

	// Get signature.
	newSignature := getSignature(signingKey, formValues.Get("Policy"))

	// Verify signature.
	if !compareSignatureV4(newSignature, formValues.Get(AmzSignature)) {
		return ErrSignatureDoesNotMatch
	}

	// Success.
	return nil
}

// VerifyPreSignature - Verify query headers with pre-signed signature
//   - http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-html
//
// returns ErrNone if the signature matches.
func (s *signer) VerifyPreSignature(hashedPayload string, r *http.Request) error {
	// Copy request
	req := *r

	// Parse request query string.
	pSignValues, err := s.parsePreSignV4(req.URL.Query())
	if err != nil {
		return err
	}

	// Extract all the signed headers along with its values.
	extractedSignedHeaders, errCode := extractSignedHeaders(pSignValues.SignedHeaders, r)
	if errCode != nil {
		return errCode
	}

	// If the host which signed the request is slightly ahead in time (by less than DefaultSkewLimit) the
	// request should still be allowed.
	if pSignValues.Date.After(time.Now().UTC().Add(s.v4.maxRequestTimeSkew)) {
		return ErrRequestNotReadyYet
	}

	if time.Now().UTC().Sub(pSignValues.Date) > pSignValues.Expires {
		return ErrExpiredPresignRequest
	}

	// Save the date and expires.
	t := pSignValues.Date
	expireSeconds := int(pSignValues.Expires / time.Second)

	// Construct new query.
	query := make(url.Values)
	clientHashedPayload := req.URL.Query().Get(AmzContentSha256)
	if clientHashedPayload != "" {
		query.Set(AmzContentSha256, hashedPayload)
	}

	token := req.URL.Query().Get(AmzSecurityToken)
	if token != "" {
		query.Set(AmzSecurityToken, s.sessionToken)
	}

	query.Set(AmzAlgorithm, signV4Algorithm)

	// Construct the query.
	query.Set(AmzDate, t.Format(iso8601Format))
	query.Set(AmzExpires, strconv.Itoa(expireSeconds))
	query.Set(AmzSignedHeaders, getSignedHeadersV4(extractedSignedHeaders))
	query.Set(AmzCredential, s.accessKey+SlashSeparator+pSignValues.Credential.getScope())

	defaultSigParams := []string{
		AmzContentSha256,
		AmzSecurityToken,
		AmzAlgorithm,
		AmzDate,
		AmzExpires,
		AmzSignedHeaders,
		AmzCredential,
		AmzSignature,
	}

	//Add missing query parameters if any provided in the request URL
	//将其它参数加入签名
	for k, v := range req.URL.Query() {
		if !lo.Contains(defaultSigParams, k) {
			if _, ok := SupportedHeadGetReqParams[strings.ToLower(k)]; ok { // 这些参数是必须要添加签名的
				query[k] = v
			} else if s.v4.validateQuery { // 如果ValidateQuery的话 则全部加入签名, 不然就不加入签名了
				query[k] = v
			}
		}
	}

	// Get the encoded query.
	encodedQuery := query.Encode()

	// Verify if date query is same.
	if req.URL.Query().Get(AmzDate) != query.Get(AmzDate) {
		return ErrSignatureDoesNotMatch
	}
	// Verify if expires query is same.
	if req.URL.Query().Get(AmzExpires) != query.Get(AmzExpires) {
		return ErrSignatureDoesNotMatch
	}
	// Verify if signed headers query is same.
	if req.URL.Query().Get(AmzSignedHeaders) != query.Get(AmzSignedHeaders) {
		return ErrSignatureDoesNotMatch
	}
	// Verify if credential query is same.
	if req.URL.Query().Get(AmzCredential) != query.Get(AmzCredential) {
		return ErrSignatureDoesNotMatch
	}
	// Verify if sha256 payload query is same.
	if clientHashedPayload != "" && clientHashedPayload != query.Get(AmzContentSha256) {
		return ErrContentSHA256Mismatch
	}
	// Verify if security token is correct.
	if token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(s.sessionToken)) != 1 {
		return ErrInvalidToken
	}

	method := r.Method

	/// Verify finally if signature is same.

	// Get canonical request.
	preSignedCanonicalReq := getCanonicalRequestV4(extractedSignedHeaders, hashedPayload, encodedQuery, req.URL.Path, method)

	// Get string to sign from canonical request.
	preSignedStringToSign := getStringToSignV4(preSignedCanonicalReq, t, pSignValues.Credential.getScope())

	// Get hmac pre-signed signing key.
	preSignedSigningKey := getSigningKeyV4(s.secretKey, pSignValues.Credential.scope.date, pSignValues.Credential.scope.region, pSignValues.Credential.scope.service)

	// Get new signature.
	newSignature := getSignature(preSignedSigningKey, preSignedStringToSign)

	// Verify signature.
	if !compareSignatureV4(req.URL.Query().Get(AmzSignature), newSignature) {
		return ErrSignatureDoesNotMatch
	}
	return nil
}

// VerifyHeadSignature - Verify authorization header with calculated header in accordance with
//   - http://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html
//
// returns ErrNone if signature matches.
func (s *signer) VerifyHeadSignature(hashedPayload string, r *http.Request) error {
	// Copy request.
	req := *r

	// Save authorization header.
	v4Auth := req.Header.Get(hAuthorization)

	// Parse signature version '4' header.
	signV4Values, err := s.parseSignV4(v4Auth)
	if err != nil {
		return err
	}

	// Extract all the signed headers along with its values.
	extractedSignedHeaders, errCode := extractSignedHeaders(signV4Values.SignedHeaders, r)
	if errCode != nil {
		return errCode
	}

	// Extract date, if not present throw error.
	var date string
	if date = req.Header.Get(AmzDate); date == "" {
		if date = r.Header.Get(hDate); date == "" {
			return ErrMissingDateHeader
		}
	}

	// Parse date header.
	t, e := time.Parse(iso8601Format, date)
	if e != nil {
		return ErrMalformedDate
	}

	// Query string.
	queryStr := req.URL.Query().Encode()

	method := r.Method

	// Get canonical request.
	canonicalRequest := getCanonicalRequestV4(extractedSignedHeaders, hashedPayload, queryStr, req.URL.Path, method)

	// Get string to sign from canonical request.
	stringToSign := getStringToSignV4(canonicalRequest, t, signV4Values.Credential.getScope())

	// Get hmac signing key.
	signingKey := getSigningKeyV4(s.secretKey, signV4Values.Credential.scope.date, signV4Values.Credential.scope.region, signV4Values.Credential.scope.service)

	// Calculate signature.
	newSignature := getSignature(signingKey, stringToSign)

	// Verify if signature match.
	if !compareSignatureV4(newSignature, signV4Values.Signature) {
		return ErrSignatureDoesNotMatch
	}

	// Return error none.
	return nil
}
