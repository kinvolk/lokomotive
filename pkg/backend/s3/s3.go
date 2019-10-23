package s3

import (
	"os"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/backend"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

type s3 struct {
	Bucket        string `hcl:"bucket"`
	Key           string `hcl:"key"`
	Region        string `hcl:"region,optional"`
	AWSCredsPath  string `hcl:"aws_creds_path,optional"`
	DynamoDBTable string `hcl:"dynamodb_table,optional"`
}

// init registers s3 as a backend
func init() {
	backend.Register("s3", NewS3Backend())
}

//Loadconfig loads configuration for s3 backend
func (s *s3) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, s)
}

func NewS3Backend() *s3 {
	return &s3{}
}

// Render renders the go template with s3 backend configuration
func (s *s3) Render() (string, error) {

	return util.RenderTemplate(backendConfigTmpl, s)
}

// Validate validates S3 backend configuration
func (s *s3) Validate() error {

	if s.Bucket == "" {
		return errors.Errorf("no bucket specified")
	}

	if s.Key == "" {
		return errors.Errorf("no key specified")
	}

	if s.AWSCredsPath == "" && os.Getenv("AWS_SHARED_CREDENTIALS_FILE") == "" {
		if s.Region == "" && os.Getenv("AWS_DEFAULT_REGION") == "" {
			return errors.Errorf("no region specified: use Region field in backend configuration or AWS_DEFAULT_REGION environment variable")
		}
	}

	return nil
}
