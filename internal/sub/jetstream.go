package sub

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type (
	JetStreamOpts struct {
		Endpoint    string
		JetStreamID string
		Logg        *slog.Logger
	}

	JetStreamSub struct {
		jsIter    jetstream.MessagesContext
		logg      *slog.Logger
		natsConn  *nats.Conn
		durableID string
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
		jsIter:    iter,
		natsConn:  natsConn,
		logg:      o.Logg,
		durableID: o.JetStreamID,
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

		s.logg.Debug("processing nats message", "subject", msg.Subject())
	}
}
