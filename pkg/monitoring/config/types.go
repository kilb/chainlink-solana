// package config parses flags, environment variables and json object to build
// a Config object that's used througout the monitor.
package config

import (
	"time"

	"github.com/gagliardetto/solana-go"
)

type Config struct {
	Solana         Solana
	Kafka          Kafka
	SchemaRegistry SchemaRegistry
	Feeds          Feeds
	Http           Http
	Feature        Feature
}

type Solana struct {
	RPCEndpoint  string
	NetworkName  string
	NetworkID    string
	ChainID      string
	ReadTimeout  time.Duration
	PollInterval time.Duration
}

type Kafka struct {
	Brokers          string
	ClientID         string
	SecurityProtocol string

	SaslMechanism string
	SaslUsername  string
	SaslPassword  string

	TransmissionTopic        string
	ConfigSetTopic           string
	ConfigSetSimplifiedTopic string
}

type SchemaRegistry struct {
	URL      string
	Username string
	Password string
}

type Feeds struct {
	// If URL is set, the RDD tracker will start and override any feed configs extracted from FilePath!
	FilePath        string
	URL             string
	RDDReadTimeout  time.Duration
	RDDPollInterval time.Duration
	Feeds           []Feed
}

type Feed struct {
	// Data extracted from the RDD
	FeedName       string
	FeedPath       string
	Symbol         string
	HeartbeatSec   int64
	ContractType   string
	ContractStatus string

	// Equivalent to ProgramID in Solana
	ContractAddress      solana.PublicKey
	TransmissionsAccount solana.PublicKey
	StateAccount         solana.PublicKey
}

type Http struct {
	Address string
}

type Feature struct {
	// If set, the monitor will not read from a chain instead from a source of random state snapshots.
	TestOnlyFakeReaders bool
	// If set, the monitor will not read from the RDD, instead it will get data from a local source of random feeds configurations.
	TestOnlyFakeRdd bool
}
