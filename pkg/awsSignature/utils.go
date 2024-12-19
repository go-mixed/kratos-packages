package awsSignature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/samber/lo"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Signature and API related constants.
const (
	signV2Algorithm = "AWS"
	signV4Algorithm = "AWS4-HMAC-SHA256"
	iso8601Format   = "20060102T150405Z"
	yyyymmdd        = "20060102"
)

// Whitelist resource list that will be used in query string for signature-V2 calculation.
//
// This list should be kept alphabetically sorted, do not hastily edit.
var resourceList = []string{
	"acl",
	"cors",
	"delete",
	"encryption",
	"legal-hold",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"requestPayment",
	"response-cache-control",
	"response-content-disposition",
	"response-content-encoding",
	"response-content-language",
	"response-content-type",
	"response-expires",
	"retention",
	"select",
	"select-type",
	"tagging",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}

// IsRequestSignatureV4 Verify if request has AWS Signature Version '4'.
func IsRequestSignatureV4(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get(hAuthorization), signV4Algorithm)
}

// IsRequestSignatureV2 Verify if request has AWS Signature Version '2'.
func IsRequestSignatureV2(r *http.Request) bool {
	return !strings.HasPrefix(r.Header.Get(hAuthorization), signV4Algorithm) &&
		strings.HasPrefix(r.Header.Get(hAuthorization), signV2Algorithm)
}

// IsRequestPreSignatureV4 Verify if request has AWS PreSign Version '4'.
func IsRequestPreSignatureV4(r *http.Request) bool {
	_, ok := r.URL.Query()["X-Amz-Credential"]
	return ok
}

// IsRequestPreSignatureV2 Verify request has AWS PreSign Version '2'.
func IsRequestPreSignatureV2(r *http.Request) bool {
	_, ok := r.URL.Query()["AWSAccessKeyId"]
	return ok
}

// IsRequestPostPolicySignatureV4 Verify if request has AWS Post policy Signature Version '4'.
func IsRequestPostPolicySignatureV4(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") &&
		r.Method == http.MethodPost
}

// IsRequestSignStreamingV4 Verify if the request has AWS Streaming Signature Version '4'. This is only valid for 'PUT' operation.
func IsRequestSignStreamingV4(r *http.Request) bool {
	return r.Header.Get(AmzContentSha256) == streamingContentSHA256 &&
		r.Method == http.MethodPut
}

// skipContentSha256Checksum returns true if caller needs to skip
// payload checksum, false if not.
func skipContentSha256Checksum(r *http.Request) bool {
	var (
		v  []string
		ok bool
	)

	if IsRequestPreSignatureV4(r) {
		v, ok = r.URL.Query()[AmzContentSha256]
		if !ok {
			v, ok = r.Header[AmzContentSha256]
		}
	} else {
		v, ok = r.Header[AmzContentSha256]
	}

	// If x-amz-content-sha256 is set and the value is not 'UNSIGNED-PAYLOAD'/"STREAMING-AWS4-HMAC-SHA256-PAYLOAD" we should validate the content sha256.
	return !(ok && v[0] != unsignedPayload && v[0] != streamingContentSHA256)
}

// GetContentSha256Checksum Returns SHA256 for calculating canonical-request.
func GetContentSha256Checksum(r *http.Request, sType string) (string, error) {
	if sType == ServiceSTS {
		payload, err := io.ReadAll(io.LimitReader(r.Body, StsRequestBodyLimit))
		if err != nil {
			return "", err
		}
		sum256 := sha256.Sum256(payload)
		r.Body = io.NopCloser(bytes.NewReader(payload))
		return hex.EncodeToString(sum256[:]), nil
	}

	var (
		defaultSha256Checksum string
		v                     []string
		ok                    bool
	)

	// For a pre-signed request we look at the query param for sha256.
	if IsRequestPreSignatureV4(r) {
		// X-Amz-Content-Sha256, if not set in pre-signed requests, checksum
		// will default to 'UNSIGNED-PAYLOAD'.
		defaultSha256Checksum = unsignedPayload
		v, ok = r.URL.Query()[AmzContentSha256]
		if !ok {
			v, ok = r.Header[AmzContentSha256]
		}
	} else {
		// X-Amz-Content-Sha256, if not set in signed requests, checksum
		// will default to sha256([]byte("")).
		defaultSha256Checksum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		v, ok = r.Header[AmzContentSha256]
	}

	// We found 'X-Amz-Content-Sha256' return the captured value.
	if ok {
		return v[0], nil
	}

	// We couldn't find 'X-Amz-Content-Sha256'.
	return defaultSha256Checksum, nil
}

// extractSignedHeaders extract signed headers from Authorization header
func extractSignedHeaders(signedHeaders []string, r *http.Request) (http.Header, error) {
	reqHeaders := r.Header
	reqQueries := r.URL.Query()
	// find whether "host" is part of list of signed headers.
	// if not return ErrUnsignedHeaders. "host" is mandatory.
	if lo.IndexOf(signedHeaders, "host") == -1 {
		return nil, ErrUnsignedHeaders
	}
	extractedSignedHeaders := make(http.Header)
	for _, header := range signedHeaders {
		// `host` will not be found in the headers, can be found in r.Host.
		// but its alway necessary that the list of signed headers containing host in it.
		val, ok := reqHeaders[http.CanonicalHeaderKey(header)]
		if !ok {
			// try to set headers from Query String
			val, ok = reqQueries[header]
		}
		if ok {
			extractedSignedHeaders[http.CanonicalHeaderKey(header)] = val
			continue
		}
		switch header {
		case "expect":
			// Golang http server strips off 'Expect' header, if the
			// client sent this as part of signed headers we need to
			// handle otherwise we would see a signature mismatch.
			// `aws-cli` sets this as part of signed headers.
			//
			// According to
			// http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.20
			// Expect header is always of form:
			//
			//   Expect       =  "Expect" ":" 1#expectation
			//   expectation  =  "100-continue" | expectation-extension
			//
			// So it safe to assume that '100-continue' is what would
			// be sent, for the time being keep this work around.
			// Adding a *TODO* to remove this later when Golang server
			// doesn't filter out the 'Expect' header.
			extractedSignedHeaders.Set(header, "100-continue")
		case "host":
			// Go http server removes "host" from Request.Header
			extractedSignedHeaders.Set(header, r.Host)
		case "transfer-encoding":
			// Go http server removes "host" from Request.Header
			extractedSignedHeaders[http.CanonicalHeaderKey(header)] = r.TransferEncoding
		case "content-length":
			// Signature-V4 spec excludes Content-Length from signed headers list for signature calculation.
			// But some clients deviate from this rule. Hence we consider Content-Length for signature
			// calculation to be compatible with such clients.
			extractedSignedHeaders.Set(header, strconv.FormatInt(r.ContentLength, 10))
		default:
			return nil, ErrUnsignedHeaders
		}
	}
	return extractedSignedHeaders, nil
}

// Trim leading and trailing spaces and replace sequential spaces with one space, following Trimall()
// in http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func signV4TrimAll(input string) string {
	// Compress adjacent spaces (a space is determined by
	// unicode.IsSpace() internally here) to one space and return
	return strings.Join(strings.Fields(input), " ")
}

// sumHMAC calculate hmac between two input byte array.
func sumHMAC(key []byte, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
