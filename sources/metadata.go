package sources

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
	"github.com/davepgreene/propsd-agent/parsers"
	"github.com/davepgreene/propsd-agent/utils"
	"encoding/json"
)

type MetadataOptions struct {
}

type MetadataChannelResponse struct {
	Path string
	Body string
}

type MetadataChannelErrorResponse struct {
	Path  string
	Error error
}

type Metadata struct {
	client *ec2metadata.EC2Metadata
	parser *parsers.Metadata
}

func NewMetadataSource(session session.Session) *Metadata {
	// We need to assemble our client manually so we can override host and timeout if we want
	c := session.ClientConfig("ec2metadata", aws.NewConfig())

	return &Metadata{
		client: utils.CreateMetadataClient(c),
		parser: parsers.NewMetadataParser(session),
	}
}

func (m *Metadata) Get() {
	resc, errc := make(chan MetadataChannelResponse), make(chan MetadataChannelErrorResponse)
	paths := map[string]func(string) (string, error){
		"instance-identity/document": m.client.GetDynamicData,
		"hostname":                   m.client.GetMetadata,
		"local-ipv4":                 m.client.GetMetadata,
		"local-hostname":             m.client.GetMetadata,
		"public-hostname":            m.client.GetMetadata,
		"public-ipv4":                m.client.GetMetadata,
		"reservation-id":             m.client.GetMetadata,
		"security-groups":            m.client.GetMetadata,
		"instance-identity/pkcs7":    m.client.GetDynamicData,
		"iam/security-credentials/":   m.client.GetMetadata,
		"network/interfaces/macs/":    m.client.GetMetadata,
	}

	for path, fn := range paths {
		go m.fetch(resc, errc, path, fn, m.parser.Parsers[path])
	}

	for i := 0; i < len(paths); i++ {
		select {
		case res := <-resc:
			log.Debugf("Parsed data from %s", res.Path)
		case err := <-errc:
			utils.AwsServiceError(m.client.ServiceName, err.Path, err.Error)
		}
	}

	// We can use goroutines for all the other metadata but because ASG relies on instance region and ID we
	// have to wait until those are complete.
	m.AutoScaling()
}

func (m *Metadata) Tags() {
	m.parser.Parsers["tags"]("")
}

func (m *Metadata) AutoScaling() {
	m.parser.Parsers["auto-scaling-group"]("")
}

func (m *Metadata) fetch(resc chan MetadataChannelResponse, errc chan MetadataChannelErrorResponse, path string, method func(string) (string, error), parser func(string)) {
	body, err := method(path)
	if err != nil {
		errc <- MetadataChannelErrorResponse{
			Path: path,
			Error: err,
		}
		return
	}
	parser(body)
	resc <- MetadataChannelResponse{
		Path: path,
		Body: body,
	}
}

func (m *Metadata) Properties() *parsers.MetadataProperties {
	return m.parser.Properties()
}

func (m *Metadata) Ok() bool {
	// Attempt to marshal the metadata object and determine its length. This
	// way we can figure out if we actually ended up with valid metadata
	// properties.
	b, err := json.Marshal(m.Properties())
	if err != nil || len(b) == 0 {
		return false
	}

	return true
}
