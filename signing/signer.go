package signing

import (
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
)

type Signer interface {
	Sign(req *http.Request, body io.ReadSeeker) error
}

func NewNoOpSigner() Signer {
	return &noOpSigner{}
}

type noOpSigner struct{}

var _ Signer = (*noOpSigner)(nil)

func (noop *noOpSigner) Sign(_ *http.Request, _ io.ReadSeeker) error {
	return nil
}

func NewAwsSigner(region string, credentials *credentials.Credentials) Signer {
	return &awsSigner{
		region: region,
		signer: v4.NewSigner(credentials),
	}
}

type awsSigner struct {
	region string
	signer *v4.Signer
}

var _ Signer = (*awsSigner)(nil)

func (aws *awsSigner) Sign(req *http.Request, body io.ReadSeeker) error {
	_, err := aws.signer.Sign(req, body, "es", aws.region, time.Now())
	return err
}
