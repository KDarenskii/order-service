package broker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/pkg/broker/codec"
)

var _ Bus[any] = (*kafkaBus[any])(nil)

const (
	consumeBackoffInitial = 500 * time.Millisecond
	consumeBackoffMax     = 30 * time.Second
	consumeMaxRetries     = 5
)

type EventIdGetter interface {
	EventId() string
}

type consumerGroupHandler[T any] struct {
	codec   codec.Codec[T]
	handler MessageHandler[T]
	topic   string
}

type kafkaBus[T any] struct {
	client        *KafkaClient
	codec         codec.Codec[T]
	topic         string
	consumerGroup string
	consumer      sarama.ConsumerGroup
}

func NewBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) (Bus[T], error) {
	if client == nil || topic == "" {
		return nil, fmt.Errorf("broker: empty client or topic configuration")
	}

	group := Coalesce(consumerGroup, client.DefaultConsumerGroup())

	if group == "" {
		return nil, fmt.Errorf("broker: empty consumer group")
	}

	bus := kafkaBus[T]{
		client:        client,
		codec:         c,
		topic:         topic,
		consumerGroup: group,
	}

	return &bus, nil
}

func MustKafkaBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) Bus[T] {
	bus, err := NewBus(client, c, topic, consumerGroup)
	if err != nil {
		log.Fatal().Err(err).Str("topic", topic).Msg("Failed to create bus")
	}

	return bus
}

func (b *kafkaBus[T]) Send(_ context.Context, msg *T, headers ...Header) error {
	encodedMsg, err := b.codec.Encode(msg)
	if err != nil {
		return fmt.Errorf("broker: failed to encode message")
	}

	eventId := getEventId(msg)

	msgHeaders := make([]sarama.RecordHeader, 0, len(headers))

	for _, header := range headers {
		msgHeaders = append(
			msgHeaders, sarama.RecordHeader{
				Key:   []byte(header.Key),
				Value: []byte(header.Value),
			},
		)
	}

	producerMsg := &sarama.ProducerMessage{
		Topic:   b.topic,
		Key:     sarama.StringEncoder(eventId),
		Value:   sarama.ByteEncoder(encodedMsg),
		Headers: msgHeaders,
	}

	_, _, err = b.client.producer.SendMessage(producerMsg)
	if err != nil {
		return fmt.Errorf("broker: failed to send message")
	}

	return nil
}

func (b *kafkaBus[T]) QueueName() string {
	return b.topic
}

func (b *kafkaBus[T]) Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error {
	consumer, err := b.client.NewConsumerGroup(b.consumerGroup)
	if err != nil {
		return fmt.Errorf("broker: create consumer group %q: %w", b.consumerGroup, err)
	}
	b.consumer = consumer

	h := &consumerGroupHandler[T]{codec: b.codec, handler: handler, topic: b.topic}

	if wg != nil {
		wg.Add(1)
	}

	go func() {
		for err := range consumer.Errors() {
			log.Warn().Err(err).Str("topic", b.topic).Msg("consumer group error")
		}
	}()

	go func() {
		if wg != nil {
			defer wg.Done()
		}
		defer func() {
			if err := consumer.Close(); err != nil && !errors.Is(err, sarama.ErrClosedClient) {
				log.Error().Err(err).Str("topic", b.topic).Msg("failed to close consumer group")
			}
		}()

		backoff := consumeBackoffInitial
		fails := 0
		for {
			err := consumer.Consume(ctx, []string{b.topic}, h)

			if ctx.Err() != nil {
				return
			}

			if err == nil {
				backoff = consumeBackoffInitial
				fails = 0
				continue
			}

			fails++

			if fails >= consumeMaxRetries {
				log.Fatal().Err(err).Str("topic", b.topic).Int(
					"attempts", consumeMaxRetries,
				).Msg("max retries exceeded")
			}

			log.Warn().Err(err).Str("topic", b.topic).Int(
				"current_attempt", fails,
			).Int("max_attempts", consumeMaxRetries).Msg("failed to consume, retrying")

			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}

			backoff = consumeBackoffMax
		}
	}()

	return nil
}

func (b *kafkaBus[T]) Close() error {
	if b.consumer != nil {
		return b.consumer.Close()
	}

	return nil
}

func getEventId[T any](v *T) string {
	if v != nil {
		if getter, ok := any(v).(EventIdGetter); ok {
			if id := getter.EventId(); id != "" {
				return id
			}
		}
	}

	return uuid.Must(uuid.NewV4()).String()
}

func (h *consumerGroupHandler[T]) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler[T]) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler[T]) ConsumeClaim(
	session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim,
) error {
	ctx := session.Context()
	defer session.Commit()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case msg, ok := <-claim.Messages():
		if !ok {
			return nil
		}

		if err := h.process(ctx, session, msg); err != nil {
			return err
		}
	}

	return nil
}

func (h *consumerGroupHandler[T]) process(
	ctx context.Context, session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage,
) error {
	decoded, err := h.codec.Decode(msg.Value)
	if err != nil {
		log.Error().Err(err).Str("topic", msg.Topic).Msg("failed to decode message")
		session.MarkMessage(msg, "")
		return nil
	}

	rawHeaders := make([]Header, 0, len(msg.Headers))

	for _, header := range msg.Headers {
		rawHeaders = append(
			rawHeaders, Header{
				Key:   string(header.Key),
				Value: string(header.Value),
			},
		)
	}

	err = h.handler(ctx, decoded, rawHeaders)

	if err != nil && IsNotCriticalError(err) {
		log.Warn().Err(err).Str("topic", msg.Topic).Msg("failed to process message, not critical error, skipping")
		session.MarkMessage(msg, "")
		return nil
	}

	if err != nil {
		log.Error().Err(err).Str("topic", msg.Topic).Msg("failed to process message, critical error")
		return err
	}

	session.MarkMessage(msg, "")

	return nil
}
