package monitoring

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-solana/pkg/monitoring/config"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// This benchmark measures how many messages end up in the kafka client given
// that the chain readers respond immediately with random data and the rdd poller
// will generate a new set of 5 random feeds every second.

//goos: darwin
//goarch: amd64
//pkg: github.com/smartcontractkit/chainlink-solana/pkg/monitoring
//cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
// (10jan2022)
//    5719	    184623 ns/op	   91745 B/op	    1482 allocs/op
func BenchmarkManager(b *testing.B) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.NewNullLogger()

	cfg := config.Config{}
	cfg.Solana.PollInterval = 0 // poll as quickly as possible.
	cfg.Feeds.URL = "http://some-fake-url-just-to-trigger-rdd-polling.com"
	cfg.Feeds.RDDPollInterval = 1 * time.Second
	cfg.Feeds.RDDReadTimeout = 1 * time.Second

	transmissionSchema := fakeSchema{transmissionCodec}
	configSetSchema := fakeSchema{configSetCodec}
	configSetSimplifiedSchema := fakeSchema{configSetCodec}

	producer := fakeProducer{make(chan producerMessage), ctx}

	transmissionReader := NewRandomDataReader(ctx, wg, "transmission", log)
	stateReader := NewRandomDataReader(ctx, wg, "state", log)

	monitor := NewMultiFeedMonitor(
		cfg.Solana,

		log,
		transmissionReader, stateReader,
		producer,
		&devnullMetrics{},

		cfg.Kafka.ConfigSetTopic,
		cfg.Kafka.ConfigSetSimplifiedTopic,
		cfg.Kafka.TransmissionTopic,

		configSetSchema,
		configSetSimplifiedSchema,
		transmissionSchema,
	)

	source := NewFakeRDDSource(5, 6) // Always produce 5 random feeds.
	rddPoller := NewSourcePoller(
		source,
		log,
		cfg.Feeds.RDDPollInterval,
		cfg.Feeds.RDDReadTimeout,
		0, // no buffering!
	)

	manager := NewManager(
		log,
		rddPoller,
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		rddPoller.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		manager.Start(ctx, wg, monitor.Start)
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		select {
		case <-producer.sendCh:
			// Drain the producer channel.
		case <-ctx.Done():
			continue
		}
	}
}
