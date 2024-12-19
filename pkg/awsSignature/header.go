package awsSignature

const (
	AmzContentSha256 = "X-Amz-Content-Sha256"
	AmzDate          = "X-Amz-Date"
	AmzAlgorithm     = "X-Amz-Algorithm"
	AmzExpires       = "X-Amz-Expires"
	AmzSignedHeaders = "X-Amz-SignedHeaders"
	AmzSignature     = "X-Amz-Signature"
	AmzCredential    = "X-Amz-Credential"
	AmzSecurityToken = "X-Amz-Security-Token"

	hExpires            = "Expires"
	hContentType        = "Content-Type"
	hCacheControl       = "Cache-Control"
	hContentEncoding    = "Content-Encoding"
	hContentLanguage    = "Content-Language"
	hContentDisposition = "Content-Disposition"

	streamingContentSHA256 = "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"
	// http Header AmzContentSha256 == "UNSIGNED-PAYLOAD" indicates that the
	// client did not calculate sha256 of the payload.
	unsignedPayload = "UNSIGNED-PAYLOAD"

	hAuthorization = "Authorization"
	hDate          = "Date"
)

// SupportedHeadGetReqParams - supported request parameters for GET and HEAD presigned request.
var SupportedHeadGetReqParams = map[string]string{
	"response-expires":             hExpires,
	"response-content-type":        hContentType,
	"response-cache-control":       hCacheControl,
	"response-content-encoding":    hContentEncoding,
	"response-content-language":    hContentLanguage,
	"response-content-disposition": hContentDisposition,
}
