package stream

import (
	"errors"
	sc "github.com/bsm/sarama-cluster"
	s "gopkg.in/Shopify/sarama.v1"
)

var (
	ErrSLInUse = errors.New("pylon/stream: this SL is in use")
)

// This is a Stream Layer Reader. Create one of these for every job in Kafka you will be doing, e.g. make one for the rack master as well as every individual Pylon.
type SLRead struct {
	Kafka  *sc.Client
	inUse  bool
	Config *Config
}

type SLWrite struct {
	Kafka  *s.Client
	inUse  bool
	Config *Config
}

type SLOptions struct {
	Addrs        []string
	ClientID     string
	SaramaConfig *sc.Config
	SLConfig     *Config
}

type Config struct {
	Topic string
}

// Creates a new "SL" or Stream Layer. The returning object can be passed into a Producer or Consumer creation for further use.
func NewSL(opts *SLOptions) (*SL, error) {
	var cfg *sc.Config
	if opts == nil {
		cfg = opts.SaramaConfig
	} else {
		cfg = sc.NewConfig()
	}

	cfg.ClientID = opts.ClientID

	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	kafka, err := sc.NewClient(opts.Addrs, cfg)
	if err != nil {
		return nil, err
	}

	return &SL{
		Kafka:  kafka,
		inUse:  false,
		Config: opts.SLConfig,
	}, nil
}

func (s *SL) Lock() error {
	if s.inUse {
		return ErrSLInUse
	}

	s.inUse = true

	return nil
}
