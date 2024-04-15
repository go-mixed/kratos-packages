package sign

import "gopkg.in/go-mixed/kratos-packages.v2/pkg/log"

type signOptions struct {
	signedFields      []string
	includeBlank      bool
	logger            *log.Helper
	validateTimestamp bool
}

func defaultOptions() *signOptions {
	return &signOptions{
		nil,
		false,
		nil,
		true,
	}
}

func (o *signOptions) apply(opts ...Option) *signOptions {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Option func(*signOptions)

func WithSignedFields(signedFields ...string) Option {
	return func(o *signOptions) {
		o.signedFields = signedFields
	}
}

func WithBlank(withBlank bool) Option {
	return func(o *signOptions) {
		o.includeBlank = withBlank
	}
}

func WithLogger(logger *log.Helper) Option {
	return func(o *signOptions) {
		o.logger = logger
	}
}

func WithValidateTimestamp(validateTimestamp bool) Option {
	return func(o *signOptions) {
		o.validateTimestamp = validateTimestamp
	}
}
