package awsSignature

import (
	"net/url"
	"strings"
	"time"
)

// credentialHeader data type represents structured form of Credential
// string from authorization header.
type credentialHeader struct {
	accessKey string
	scope     struct {
		date    time.Time
		region  string
		service string
		request string
	}
}

// Return scope string.
func (c credentialHeader) getScope() string {
	return strings.Join([]string{
		c.scope.date.Format(yyyymmdd),
		c.scope.region,
		c.scope.service,
		c.scope.request,
	}, SlashSeparator)
}

// parse credentialHeader string into its structured form.
// eg: Credential=admin/20210601/us-east-1/s3/aws4_request
func parseCredentialHeaderV4(credElement string) (ch credentialHeader, aec error) {
	creds := strings.SplitN(strings.TrimSpace(credElement), "=", 2)
	if len(creds) != 2 {
		return ch, ErrMissingFields
	}

	if creds[0] != "Credential" {
		return ch, ErrMissingCredTag
	}

	credElements := strings.Split(strings.TrimSpace(creds[1]), SlashSeparator)
	if len(credElements) < 5 {
		return ch, ErrCredMalformed
	}

	// accessKey 允许有 /, 将其合并起来
	accessKey := strings.Join(credElements[:len(credElements)-4], SlashSeparator) // The access key may contain one or more `/`
	if accessKey == "" {
		return ch, ErrInvalidAccessKeyID
	}

	// Save access key id.
	cred := credentialHeader{
		accessKey: accessKey,
	}

	credElements = credElements[len(credElements)-4:]
	var e error
	cred.scope.date, e = time.Parse(yyyymmdd, credElements[0])
	if e != nil {
		return ch, ErrMalformedCredentialDate
	}

	cred.scope.region = credElements[1]

	if credElements[2] != ServiceSTS && credElements[2] != ServiceS3 {
		return ch, ErrInvalidServiceName
	}
	cred.scope.service = credElements[2]
	if credElements[3] != "aws4_request" {
		return ch, ErrInvalidRequestVersion
	}
	cred.scope.request = credElements[3]
	return cred, nil
}

// Parse signature from signature tag.
// eg: Signature=ba40514b6a76394523d8c6605da587ee1f415173aa4846e54c16d9a4ebe66f05
func parseSignature(signElement string) (string, error) {
	signFields := strings.Split(strings.TrimSpace(signElement), "=")
	if len(signFields) != 2 {
		return "", ErrMissingFields
	}
	if signFields[0] != "Signature" {
		return "", ErrMissingSignTag
	}
	if signFields[1] == "" {
		return "", ErrMissingFields
	}
	signature := signFields[1]
	return signature, nil
}

// Parse slice of signed headers from signed headers tag.
// eg: SignedHeaders=host;x-amz-content-sha256;x-amz-date
func parseSignedHeader(signedHdrElement string) ([]string, error) {
	signedHdrFields := strings.Split(strings.TrimSpace(signedHdrElement), "=")
	if len(signedHdrFields) != 2 {
		return nil, ErrMissingFields
	}
	if signedHdrFields[0] != "SignedHeaders" {
		return nil, ErrMissingSignHeadersTag
	}
	if signedHdrFields[1] == "" {
		return nil, ErrMissingFields
	}
	signedHeaders := strings.Split(signedHdrFields[1], ";")
	return signedHeaders, nil
}

// SignValues data type represents structured form of AWS Signature V4 header.
type SignValues struct {
	Credential    credentialHeader
	SignedHeaders []string
	Signature     string
}

// preSignValues data type represents structued form of AWS Signature V4 query string.
type preSignValues struct {
	SignValues
	Date    time.Time
	Expires time.Duration
}

// IsPreSignParamsExistV4 Parses signature version '4' query string of the following form.
//
//	querystring = X-Amz-Algorithm=algorithm
//	querystring += &X-Amz-Credential= urlencode(accessKey + '/' + credential_scope)
//	querystring += &X-Amz-Date=date
//	querystring += &X-Amz-Expires=timeout interval
//	querystring += &X-Amz-SignedHeaders=signed_headers
//	querystring += &X-Amz-Signature=signature
//
// verifies if any of the necessary query params are missing in the presigned request.
func IsPreSignParamsExistV4(query url.Values) error {
	v4PreSignQueryParams := []string{AmzAlgorithm, AmzCredential, AmzSignature, AmzDate, AmzSignedHeaders, AmzExpires}
	for _, v4PreSignQueryParam := range v4PreSignQueryParams {
		if _, ok := query[v4PreSignQueryParam]; !ok {
			return ErrInvalidQueryParams
		}
	}
	return nil
}

// Parses all the pre-signed signature values into separate elements.
func (s *signer) parsePreSignV4(query url.Values) (psv preSignValues, aec error) {
	// verify whether the required query params exist.
	aec = IsPreSignParamsExistV4(query)
	if aec != nil {
		return psv, aec
	}

	// Verify if the query algorithm is supported or not.
	if query.Get(AmzAlgorithm) != signV4Algorithm {
		return psv, ErrInvalidQuerySignatureAlgo
	}

	// Initialize signature version '4' structured header.
	preSignV4Values := preSignValues{}

	// Save credential.
	preSignV4Values.Credential, aec = parseCredentialHeaderV4("Credential=" + query.Get(AmzCredential))
	if aec != nil {
		return psv, aec
	} else if !s.v4.isValidRegion(preSignV4Values.Credential.scope.region) { // Should validate region, only if region is set.
		return preSignV4Values, ErrInvalidRegion
	}

	var e error
	// Save date in native time.Time.
	preSignV4Values.Date, e = time.Parse(iso8601Format, query.Get(AmzDate))
	if e != nil {
		return psv, ErrMalformedPresignedDate
	}

	// Save expires in native time.Duration.
	preSignV4Values.Expires, e = time.ParseDuration(query.Get(AmzExpires) + "s")
	if e != nil {
		return psv, ErrMalformedExpires
	}

	if preSignV4Values.Expires < 0 {
		return psv, ErrNegativeExpires
	}

	// Check if Expiry time is less than 7 days (value in seconds).
	if preSignV4Values.Expires.Seconds() > 604800 {
		return psv, ErrMaximumExpires
	}

	// Save signed headers.
	preSignV4Values.SignedHeaders, aec = parseSignedHeader("SignedHeaders=" + query.Get(AmzSignedHeaders))
	if aec != nil {
		return psv, aec
	}

	// Save signature.
	preSignV4Values.Signature, aec = parseSignature("Signature=" + query.Get(AmzSignature))
	if aec != nil {
		return psv, aec
	}

	// Return structed form of signature query string.
	return preSignV4Values, nil
}

// parseSignV4 Parses signature version '4' header of the following form.
//
//	Authorization: algorithm Credential=accessKeyID/credScope, \
//	        SignedHeaders=signedHeaders, Signature=signature
//
// eg: AWS4-HMAC-SHA256 Credential=admin/20210601/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=ba40514b6a76394523d8c6605da587ee1f415173aa4846e54c16d9a4ebe66f05
func (s *signer) parseSignV4(v4Auth string) (sv SignValues, aec error) {
	// credElement is fetched first to skip replacing the space in access key.
	credElement := strings.TrimPrefix(strings.Split(strings.TrimSpace(v4Auth), ",")[0], signV4Algorithm)
	// Replace all spaced strings, some clients can send spaced
	// parameters and some won't. So we pro-actively remove any spaces
	// to make parsing easier.
	v4Auth = strings.Replace(v4Auth, " ", "", -1)
	if v4Auth == "" {
		return sv, ErrAuthHeaderEmpty
	}

	// Verify if the header algorithm is supported or not.
	if !strings.HasPrefix(v4Auth, signV4Algorithm) {
		return sv, ErrSignatureVersionNotSupported
	}

	// Strip off the Algorithm prefix.
	v4Auth = strings.TrimPrefix(v4Auth, signV4Algorithm)
	authFields := strings.Split(strings.TrimSpace(v4Auth), ",")
	if len(authFields) != 3 {
		return sv, ErrMissingFields
	}

	// Initialize signature version '4' structured header.
	signV4Values := SignValues{}

	var s3Err error
	// Save credential values.
	signV4Values.Credential, s3Err = parseCredentialHeaderV4(strings.TrimSpace(credElement))
	if s3Err != nil {
		return sv, s3Err
	} else if !s.v4.isValidRegion(signV4Values.Credential.scope.region) { // Should validate region, only if region is set.
		return sv, ErrInvalidRegion
	}

	// Save signed headers.
	signV4Values.SignedHeaders, s3Err = parseSignedHeader(authFields[1])
	if s3Err != nil {
		return sv, s3Err
	}

	// Save signature.
	signV4Values.Signature, s3Err = parseSignature(authFields[2])
	if s3Err != nil {
		return sv, s3Err
	}

	// Return the structure here.
	return signV4Values, nil
}
