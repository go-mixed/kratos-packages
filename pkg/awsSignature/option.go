package awsSignature

import "time"

type option func(o *AwsSignatureV4)

func WithRegion(regions ...string) option {
	return func(o *AwsSignatureV4) {
		o.regions = regions
	}
}

func WithService(service string) option {
	return func(o *AwsSignatureV4) {
		o.serviceName = service
	}
}

func WithValidateQuery(value bool) option {
	return func(o *AwsSignatureV4) {
		o.validateQuery = value
	}
}

func WithMaxRequestTimeSkew(value time.Duration) option {
	return func(o *AwsSignatureV4) {
		o.maxRequestTimeSkew = value
	}
}
