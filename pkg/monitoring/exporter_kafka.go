package monitoring

import (
	"context"

	"github.com/smartcontractkit/chainlink-solana/pkg/monitoring/config"
	"github.com/smartcontractkit/chainlink/core/logger"
)

func NewKafkaExporter(
	solanaConfig config.Solana,
	feedConfig config.Feed,

	log logger.Logger,
	producer Producer,

	configSetSchema Schema,
	configSetSimplifiedSchema Schema,
	transmissionSchema Schema,

	configSetTopic string,
	configSetSimplifiedTopic string,
	transmissionTopic string,
) Exporter {
	return &kafkaExporter{
		solanaConfig,
		feedConfig,

		log,
		producer,

		configSetSchema,
		configSetSimplifiedSchema,
		transmissionSchema,

		configSetTopic,
		configSetSimplifiedTopic,
		transmissionTopic,
	}
}

type kafkaExporter struct {
	solanaConfig config.Solana
	feedConfig   config.Feed

	log      logger.Logger
	producer Producer

	configSetSchema           Schema
	configSetSimplifiedSchema Schema
	transmissionSchema        Schema

	configSetTopic           string
	configSetSimplifiedTopic string
	transmissionTopic        string
}

func (k *kafkaExporter) Export(ctx context.Context, data interface{}) {
	key := k.feedConfig.StateAccount.Bytes()
	switch typed := data.(type) {
	case StateEnvelope:
		func() {
			configSetMapping, err := MakeConfigSetMapping(typed, k.solanaConfig, k.feedConfig)
			if err != nil {
				k.log.Errorw("failed to map config_set", "error", err)
				return
			}
			configSetEncoded, err := k.configSetSchema.Encode(configSetMapping)
			if err != nil {
				k.log.Errorw("failed to encode config_set to Avro", "payload", configSetMapping, "error", err)
				return
			}
			if err := k.producer.Produce(key, configSetEncoded, k.configSetTopic); err != nil {
				k.log.Errorw("failed to publish config_set", "payload", configSetMapping, "error", err)
				return
			}
		}()

		func() {
			configSetSimplifiedMapping, err := MakeConfigSetSimplifiedMapping(typed, k.feedConfig)
			if err != nil {
				k.log.Errorw("failed to map config_set_simplified", "error", err)
				return
			}
			configSetSimplifiedEncoded, err := k.configSetSimplifiedSchema.Encode(configSetSimplifiedMapping)
			if err != nil {
				k.log.Errorw("failed to encode config_set_simplified to Avro", "payload", configSetSimplifiedMapping, "error", err)
				return
			}
			if err := k.producer.Produce(key, configSetSimplifiedEncoded, k.configSetSimplifiedTopic); err != nil {
				k.log.Errorw("failed to publish config_set_simplified", "payload", configSetSimplifiedMapping, "error", err)
				return
			}
		}()
	case TransmissionEnvelope:
		transmissionMapping, err := MakeTransmissionMapping(typed, k.solanaConfig, k.feedConfig)
		if err != nil {
			k.log.Errorw("failed to map transmission", "error", err)
			return
		}
		transmissionEncoded, err := k.transmissionSchema.Encode(transmissionMapping)
		if err != nil {
			k.log.Errorw("failed to encode transmission to Avro", "payload", transmissionMapping, "error", err)
			return
		}
		if err := k.producer.Produce(key, transmissionEncoded, k.transmissionTopic); err != nil {
			k.log.Errorw("failed to publish transmission", "payload", transmissionMapping, "error", err)
			return
		}
	default:
		k.log.Errorf("unknown type %T to export", data)
	}
}
