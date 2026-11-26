package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"avito/internal/handler/kafka/dto"
	"avito/internal/handler/kafka/mocks"
	"avito/internal/model"
	"avito/pkg"

	"github.com/IBM/sarama"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSession struct {
	ctx context.Context

	mu     sync.Mutex
	marked []*sarama.ConsumerMessage
}

func (s *fakeSession) Claims() map[string][]int32                       { return nil }
func (s *fakeSession) MemberID() string                                 { return "member-1" }
func (s *fakeSession) GenerationID() int32                              { return 1 }
func (s *fakeSession) MarkOffset(_ string, _ int32, _ int64, _ string)  {}
func (s *fakeSession) Commit()                                          {}
func (s *fakeSession) ResetOffset(_ string, _ int32, _ int64, _ string) {}
func (s *fakeSession) Context() context.Context                         { return s.ctx }
func (s *fakeSession) MarkMessage(msg *sarama.ConsumerMessage, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.marked = append(s.marked, msg)
}
func (s *fakeSession) markedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.marked)
}

type fakeClaim struct {
	ch chan *sarama.ConsumerMessage
}

func (c *fakeClaim) Topic() string                            { return "orders" }
func (c *fakeClaim) Partition() int32                         { return 0 }
func (c *fakeClaim) InitialOffset() int64                     { return 0 }
func (c *fakeClaim) HighWaterMarkOffset() int64               { return 0 }
func (c *fakeClaim) Messages() <-chan *sarama.ConsumerMessage { return c.ch }

func TestKafkaHandler_ConsumeClaim_Success_MarksMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ev := mocks.NewMockhandlerEvent(ctrl)
	h := NewKafkaHandler(ev, pkg.NewLogger(0))

	sess := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, 1)}

	dtoOrder := dto.Order{
		ID:        "order-1",
		CreatedAt: time.Now().UTC(),
		Status:    "created",
	}
	b, err := json.Marshal(dtoOrder)
	require.NoError(t, err)

	ev.EXPECT().
		Handle(gomock.Any(), gomock.Eq(&model.Order{
			ID:        dtoOrder.ID,
			CreatedAt: dtoOrder.CreatedAt,
			Status:    dtoOrder.Status,
		})).
		Return(nil)

	msg := &sarama.ConsumerMessage{Value: b}
	claim.ch <- msg
	close(claim.ch)

	err = h.ConsumeClaim(sess, claim)
	require.NoError(t, err)

	assert.Equal(t, 1, sess.markedCount())
}

func TestKafkaHandler_ConsumeClaim_UnmarshalError_SkipsMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ev := mocks.NewMockhandlerEvent(ctrl)
	h := NewKafkaHandler(ev, pkg.NewLogger(0))

	sess := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, 1)}

	ev.EXPECT().Handle(gomock.Any(), gomock.Any()).Times(0)

	claim.ch <- &sarama.ConsumerMessage{Value: []byte("{bad json")}
	close(claim.ch)

	err := h.ConsumeClaim(sess, claim)
	require.NoError(t, err)

	assert.Equal(t, 0, sess.markedCount())
}

func TestKafkaHandler_ConsumeClaim_HandleCanceled_ReturnsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ev := mocks.NewMockhandlerEvent(ctrl)
	h := NewKafkaHandler(ev, pkg.NewLogger(0))

	sess := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, 1)}

	dtoOrder := dto.Order{
		ID:        "order-1",
		CreatedAt: time.Now().UTC(),
		Status:    "created",
	}
	b, err := json.Marshal(dtoOrder)
	require.NoError(t, err)

	ev.EXPECT().
		Handle(gomock.Any(), gomock.Any()).
		Return(context.Canceled).
		Times(1)

	claim.ch <- &sarama.ConsumerMessage{Value: b}
	close(claim.ch)

	err = h.ConsumeClaim(sess, claim)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))

	assert.Equal(t, 0, sess.markedCount())
}

func TestKafkaHandler_ConsumeClaim_HandleOtherError_ContinuesAndMarksNext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ev := mocks.NewMockhandlerEvent(ctrl)
	h := NewKafkaHandler(ev, pkg.NewLogger(0))

	sess := &fakeSession{ctx: context.Background()}
	claim := &fakeClaim{ch: make(chan *sarama.ConsumerMessage, 2)}

	dto1 := dto.Order{ID: "order-1", CreatedAt: time.Now().UTC(), Status: "created"}
	dto2 := dto.Order{ID: "order-2", CreatedAt: time.Now().UTC().Add(time.Minute), Status: "paid"}

	b1, err := json.Marshal(dto1)
	require.NoError(t, err)
	b2, err := json.Marshal(dto2)
	require.NoError(t, err)

	gomock.InOrder(
		ev.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(errors.New("temporary")).Times(1),
		ev.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(nil).Times(1),
	)

	claim.ch <- &sarama.ConsumerMessage{Value: b1}
	claim.ch <- &sarama.ConsumerMessage{Value: b2}
	close(claim.ch)

	err = h.ConsumeClaim(sess, claim)
	require.NoError(t, err)

	assert.Equal(t, 1, sess.markedCount())
}
