package kafka

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/IBM/sarama"

	"avito/internal/handler/kafka/dto"
	"avito/internal/model"
	"avito/pkg"
)

type KafkaHandler struct {
	eventHandler handlerEvent
	log          pkg.Logger
}

func NewKafkaHandler(event handlerEvent, log pkg.Logger) *KafkaHandler {
	return &KafkaHandler{
		eventHandler: event,
		log:          log,
	}
}

func (h *KafkaHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *KafkaHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *KafkaHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for dtoMsg := range claim.Messages() {
		ctx := sess.Context()

		var order dto.Order

		if err := json.Unmarshal(dtoMsg.Value, &order); err != nil {
			h.log.Error("Error unmarshalling order %v", err)
			continue
		}

		err := h.eventHandler.Handle(ctx, &model.Order{
			ID:        order.ID,
			CreatedAt: order.CreatedAt,
			Status:    order.Status,
		})
		if err != nil {
			h.log.Error("Error processing order %v", err)
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
		sess.MarkMessage(dtoMsg, "")
	}
	return nil
}
