package sub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/pretium-ramp/internal/worker"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/riverqueue/river"
)

type (
	JetStreamOpts struct {
		Endpoint    string
		JetStreamID string
		Logg        *slog.Logger
		QueueClient *river.Client[pgx.Tx]
	}

	JetStreamSub struct {
		jsIter      jetstream.MessagesContext
		logg        *slog.Logger
		natsConn    *nats.Conn
		durableID   string
		queueClient *river.Client[pgx.Tx]
	}

	// TrackerMessage represents the NATS message payload
	TrackerMessage struct {
		InitiatorAddress string `json:"-"`
		TransactionHash  string `json:"transactionHash"`
		Amount           string `json:"amount"`
		TokenAddress     string `json:"tokenAddress"`
	}
)

const (
	pullStream  = "TRACKER"
	pullSubject = "TRACKER.*"
)

func NewJetStreamSub(o JetStreamOpts) (*JetStreamSub, error) {
	natsConn, err := nats.Connect(o.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(natsConn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := js.Stream(ctx, pullStream)
	if err != nil {
		return nil, err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       o.JetStreamID,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: pullSubject,
	})
	if err != nil {
		return nil, err
	}
	o.Logg.Info("successfully connected to NATS server")

	iter, err := consumer.Messages(
		jetstream.WithMessagesErrOnMissingHeartbeat(false),
		jetstream.PullMaxMessages(10),
	)
	if err != nil {
		return nil, err
	}

	return &JetStreamSub{
		jsIter:      iter,
		natsConn:    natsConn,
		logg:        o.Logg,
		durableID:   o.JetStreamID,
		queueClient: o.QueueClient,
	}, nil
}

func (s *JetStreamSub) Close() {
	s.jsIter.Stop()
}

func (s *JetStreamSub) Process() {
	for {
		msg, err := s.jsIter.Next()
		if err != nil {
			if errors.Is(err, jetstream.ErrMsgIteratorClosed) {
				s.logg.Debug("jetstream: iterator closed")
				return
			} else {
				s.logg.Debug("jetstream: unknown iterator error", "error", err)
				continue
			}
		}

		var chainEvent event.Event
		if err := json.Unmarshal(msg.Data(), &chainEvent); err != nil {
			s.logg.Error("failed to unmarshal chain event", "error", err)
			msg.Nak()
			continue
		}
		_, err = s.queueClient.Insert(context.Background(), worker.OfframpArgs{
			InitiatorAddress: chainEvent.Payload["from"].(string),
			TransactionHash:  chainEvent.TxHash,
			Amount:           chainEvent.Payload["value"].(string),
			TokenAddress:     chainEvent.ContractAddress,
		}, nil)

		if err != nil {
			s.logg.Error("failed to queue offramp job", "error", err)
			msg.Nak()
			continue
		}

		if err := msg.Ack(); err != nil {
			s.logg.Error("failed to ack message", "error", err)
		}

		s.logg.Info("offramp job queued successfully",
			"initiatorAddress", chainEvent.Payload["from"].(string),
			"txHash", chainEvent.TxHash)
	}
}
