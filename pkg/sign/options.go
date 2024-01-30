package sign

import "gopkg.in/go-mixed/kratos-packages.v2/pkg/log"

type Options struct {
	signedFields      []string
	withBlank         bool
	logger            *log.Helper
	validateTimestamp bool
}

func DefaultOptions() Options {
	return Options{
		nil,
		false,
		nil,
		true,
	}
}

func (o Options) WithSignedFields(signedFields ...string) Options {
	o.signedFields = signedFields
	return o
}

func (o Options) WithWithBlank(withBlank bool) Options {
	o.withBlank = withBlank
	return o
}

func (o Options) WithLogger(logger *log.Helper) Options {
	o.logger = logger
	return o
}

func (o Options) WithValidateTimestamp(validateTimestamp bool) Options {
	o.validateTimestamp = validateTimestamp
	return o
}
