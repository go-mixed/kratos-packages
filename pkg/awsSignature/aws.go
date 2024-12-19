package awsSignature

import (
	"context"
	"github.com/samber/lo"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/log"
	"net/http"
	"strings"
	"time"
)

const (
	ServiceS3      = "s3"
	ServiceSTS     = "sts"
	SlashSeparator = "/"
)

const StsRequestBodyLimit = 10 * (1 << 20) // 10 MiB.

type AwsSignatureV4 struct {
	logger *log.Helper

	maxRequestTimeSkew time.Duration
	regions            []string
	serviceName        string
	validateQuery      bool
}

type signer struct {
	v4           *AwsSignatureV4
	accessKey    string
	secretKey    string
	sessionToken string
}

func NewAwsSignatureV4(logger log.Logger, opts ...option) *AwsSignatureV4 {
	obj := &AwsSignatureV4{
		logger:             log.NewModuleHelper(logger, "auth"),
		validateQuery:      true,
		maxRequestTimeSkew: 1 * time.Hour,
		serviceName:        ServiceS3,
	}

	for _, opt := range opts {
		opt(obj)
	}
	return obj
}

// isValidRegion - verify if incoming region value is valid with configured Region.
func (s *AwsSignatureV4) isValidRegion(reqRegion string) bool {
	return lo.Contains(s.regions, reqRegion)
}

// GetAccessKeyFrom get accessKey from header
func (s *AwsSignatureV4) GetAccessKeyFrom(r *http.Request) (string, error) {
	ch, s3Err := parseCredentialHeaderV4("Credential=" + r.URL.Query().Get(AmzCredential))
	if s3Err != nil {
		// Strip off the Algorithm prefix.
		v4Auth := strings.TrimPrefix(r.Header.Get("Authorization"), signV4Algorithm)
		authFields := strings.Split(strings.TrimSpace(v4Auth), ",")
		if len(authFields) != 3 {
			return "", ErrMissingFields
		}
		ch, s3Err = parseCredentialHeaderV4(authFields[0])
		if s3Err != nil {
			return "", s3Err
		}
	}
	return ch.accessKey, nil
}

func (s *AwsSignatureV4) Signer(accessKey, secretKey, sessionToken string) *signer {
	return &signer{
		v4:           s,
		accessKey:    accessKey,
		secretKey:    secretKey,
		sessionToken: sessionToken,
	}
}

// Verify 只支持v4签名
func (s *signer) Verify(ctx context.Context, r *http.Request) error {
	sha256sum, err := GetContentSha256Checksum(r, s.v4.serviceName)
	if err != nil {
		s.v4.logger.WithContext(ctx).Error(err.Error())
	}
	switch {
	case IsRequestSignatureV4(r):
		return s.VerifyHeadSignature(sha256sum, r)
	case IsRequestPreSignatureV4(r):
		return s.VerifyPreSignature(sha256sum, r)
	default:
		return ErrInvalidToken
	}
}
