package dispatcher

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ---- Test event types ----

type testProductCreatedEvent struct {
	ProductID string `json:"product_id"`
	Title     string `json:"title"`
}

type testProductUpdatedEvent struct {
	ProductID string `json:"product_id"`
	Title     string `json:"title"`
}

type testProductDeletedEvent struct {
	ProductID string `json:"product_id"`
}

// ---- Helpers ----

// fakeLogger implements the logger.Logger interface for testing.
// It is in the dispatcher package so "logger.Logger" refers to the imported
// github.com/.../logger.Logger type. With must return that exact interface type.
type fakeLogger struct {
	mu sync.Mutex
	logs []string
	kvs [][]any
}

func newFakeLogger() *fakeLogger { return &fakeLogger{} }

func (l *fakeLogger) Debug(msg string, kv ...any) { l.record(msg, kv) }
func (l *fakeLogger) Info(msg string, kv ...any)  { l.record(msg, kv) }
func (l *fakeLogger) Warn(msg string, kv ...any)  { l.record(msg, kv) }
func (l *fakeLogger) Error(msg string, kv ...any) { l.record(msg, kv) }
func (l *fakeLogger) Fatal(msg string, kv ...any) { l.record(msg, kv) }
func (l *fakeLogger) With(kv ...any) logger.Logger { return l }

func (l *fakeLogger) record(msg string, kv []any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, msg)
	l.kvs = append(l.kvs, kv)
}


func buildEnvelope(routingKey string, payload any) []byte {
	env := types.NewEnvelope(routingKey, payload, "")
	msg, _ := types.NewRabbitMQMessage(env)
	return msg.Body
}

func makeDelivery(routingKey string, payload any) *types.DeliveryMessage {
	body := buildEnvelope(routingKey, payload)
	raw := &amqp.Delivery{
		Body:       body,
		RoutingKey:  routingKey,
		Exchange:    "test.events",
		DeliveryTag: 1,
	}
	return types.NewDeliveryMessage(raw)
}

// ---- patternMatch tests ----

func TestPatternMatch_exact(t *testing.T) {
	if !patternMatch("catalog.product.created", "catalog.product.created") {
		t.Error("expected exact match to return true")
	}
	if patternMatch("catalog.product.created", "catalog.product.deleted") {
		t.Error("expected non-matching exact to return false")
	}
}

func TestPatternMatch_singleWildcard(t *testing.T) {
	if !patternMatch("catalog.product.created", "catalog.product.*") {
		t.Error("expected catalog.product.* to match catalog.product.created")
	}
	if !patternMatch("catalog.product.deleted", "catalog.product.*") {
		t.Error("expected catalog.product.* to match catalog.product.deleted")
	}
	if patternMatch("catalog.order.created", "catalog.product.*") {
		t.Error("expected catalog.product.* to NOT match catalog.order.created")
	}
}

func TestPatternMatch_middleWildcard(t *testing.T) {
	if !patternMatch("catalog.product.created", "catalog.*.created") {
		t.Error("expected catalog.*.created to match catalog.product.created")
	}
	if !patternMatch("catalog.order.created", "catalog.*.created") {
		t.Error("expected catalog.*.created to match catalog.order.created")
	}
	if patternMatch("catalog.product.deleted", "catalog.*.created") {
		t.Error("expected catalog.*.created to NOT match catalog.product.deleted")
	}
}

func TestPatternMatch_multipleWildcards(t *testing.T) {
	if !patternMatch("a.b.c", "*.*.*") {
		t.Error("expected *.*.* to match a.b.c")
	}
	if patternMatch("a.b", "*.*.*") {
		t.Error("expected *.*.* to NOT match a.b (different segment count)")
	}
}

func TestPatternMatch_segmentCountMismatch(t *testing.T) {
	if patternMatch("catalog.product.created", "catalog.product") {
		t.Error("expected segment count mismatch to return false")
	}
	if patternMatch("catalog.product", "catalog.product.created") {
		t.Error("expected segment count mismatch to return false")
	}
}

// ---- EventDispatcher exact-match routing ----

func TestEventDispatcher_exactMatch_routesToCorrectHandler(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	var created, updated bool

	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		created = true
		return nil
	}))
	d.register("catalog.product.updated", NewTypedHandler(func(ctx context.Context, ev testProductUpdatedEvent) error {
		updated = true
		return nil
	}))

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1", Title: "Bike"})
	if err := d.Handle(ctx, msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !created {
		t.Error("catalog.product.created handler was not called")
	}
	if updated {
		t.Error("catalog.product.updated handler was incorrectly called")
	}
}

func TestEventDispatcher_exactMatch_passesEventPayload(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		if ev.ProductID != "p99" || ev.Title != "Laptop" {
			t.Errorf("unexpected event values: ProductID=%s, Title=%s", ev.ProductID, ev.Title)
		}
		return nil
	}))

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p99", Title: "Laptop"})
	if err := d.Handle(ctx, msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
}

// ---- EventDispatcher wildcard routing ----

func TestEventDispatcher_wildcardMatchesCreated(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	var called bool
	d := NewEventDispatcher(log, nil)
	d.registerWildcard("catalog.product.*", NewTypedHandler(func(ctx context.Context, ev any) error {
		called = true
		return nil
	}))

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	if err := d.Handle(ctx, msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !called {
		t.Error("wildcard handler was not called for catalog.product.created")
	}
}

func TestEventDispatcher_wildcardMatchesDeleted(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	var called bool
	d := NewEventDispatcher(log, nil)
	d.registerWildcard("catalog.product.*", NewTypedHandler(func(ctx context.Context, ev any) error {
		called = true
		return nil
	}))

	msg := makeDelivery("catalog.product.deleted", testProductDeletedEvent{ProductID: "p1"})
	if err := d.Handle(ctx, msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !called {
		t.Error("wildcard handler was not called for catalog.product.deleted")
	}
}

func TestEventDispatcher_exactMatch_takesPrecedenceOverWildcard(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	var exactCalled, wildcardCalled bool
	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		exactCalled = true
		return nil
	}))
	d.registerWildcard("catalog.product.*", NewTypedHandler(func(ctx context.Context, ev any) error {
		wildcardCalled = true
		return nil
	}))

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	if err := d.Handle(ctx, msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if !exactCalled {
		t.Error("exact-match handler was not called")
	}
	if wildcardCalled {
		t.Error("wildcard handler should not be called when exact match exists")
	}
}

// ---- No handler found ----

func TestEventDispatcher_noHandlerAckWithWarning(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	d := NewEventDispatcher(log, nil)
	// No handlers registered

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	err := d.Handle(ctx, msg)
	if err != nil {
		t.Errorf("expected no error for unregistered key, got: %v", err)
	}
}

func TestEventDispatcher_unknownKeyNoHandlerAckWithWarning(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	}))

	// "catalog.order.placed" has no handler — should ack silently
	msg := makeDelivery("catalog.order.placed", map[string]any{"order_id": "o1"})
	err := d.Handle(ctx, msg)
	if err != nil {
		t.Errorf("expected no error for unknown routing key, got: %v", err)
	}
}

// ---- Handler error + retry ----

func TestEventDispatcher_handlerErrorReturnsError(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	wantErr := errors.New("boom")
	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		return wantErr
	}))

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	err := d.Handle(ctx, msg)
	if err == nil {
		t.Fatal("expected error from handler, got nil")
	}
	// retry-go wraps errors; check the original message is preserved
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected error containing 'boom', got: %v", err)
	}
}

// ---- Decode failure ----

func TestEventDispatcher_decodeFailureReturnsError(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	}))

	// Malformed JSON body — decode will fail
	raw := &amqp.Delivery{
		Body:       []byte(`{invalid json`),
		RoutingKey:  "catalog.product.created",
		Exchange:    "test.events",
		DeliveryTag: 1,
	}
	msg := types.NewDeliveryMessage(raw)
	err := d.Handle(ctx, msg)
	if err == nil {
		t.Fatal("expected error from decode failure, got nil")
	}
}

// ---- Concurrent Handle calls ----

func TestEventDispatcher_concurrentHandle(t *testing.T) {
	log := newFakeLogger()
	ctx := context.Background()

	var mu sync.Mutex
	counts := make(map[string]int)

	d := NewEventDispatcher(log, nil)
	d.register("catalog.product.created", NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		mu.Lock()
		counts["created"]++
		mu.Unlock()
		return nil
	}))
	d.register("catalog.product.updated", NewTypedHandler(func(ctx context.Context, ev testProductUpdatedEvent) error {
		mu.Lock()
		counts["updated"]++
		mu.Unlock()
		return nil
	}))

	const n = 50
	var wg sync.WaitGroup
	wg.Add(n * 2)

	for range n {
		go func() {
			defer wg.Done()
			msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
			_ = d.Handle(ctx, msg)
		}()
		go func() {
			defer wg.Done()
			msg := makeDelivery("catalog.product.updated", testProductUpdatedEvent{ProductID: "p1"})
			_ = d.Handle(ctx, msg)
		}()
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if counts["created"] != n || counts["updated"] != n {
		t.Errorf("expected counts created=%d, updated=%d, got created=%d, updated=%d",
			n, n, counts["created"], counts["updated"])
	}
}

// ---- Builder tests ----

func TestBuilder_Build_success(t *testing.T) {
	log := newFakeLogger()
	b := NewBuilder(log)
	Register(b, "catalog.product.created", func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	})
	RegisterWildcard(b, "catalog.product.*", func(ctx context.Context, ev any) error {
		return nil
	})

	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}
	if d == nil {
		t.Fatal("Build() returned nil dispatcher")
	}

	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	if err := d.Handle(context.Background(), msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
}

func TestBuilder_Build_emptyReturnsError(t *testing.T) {
	log := newFakeLogger()
	b := NewBuilder(log)
	_, err := b.Build()
	if err == nil {
		t.Error("expected error for empty builder")
	}
}

func TestBuilder_Register_nilHandlerPanics(t *testing.T) {
	log := newFakeLogger()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil handler")
		}
	}()
	b := NewBuilder(log)
	Register[any](b, "catalog.product.created", nil)
}

func TestBuilder_Register_emptyKeyPanics(t *testing.T) {
	log := newFakeLogger()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty routing key")
		}
	}()
	b := NewBuilder(log)
	Register(b, "", func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	})
}

func TestBuilder_RegisterWildcard_noAsteriskPanics(t *testing.T) {
	log := newFakeLogger()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for pattern without asterisk")
		}
	}()
	b := NewBuilder(log)
	RegisterWildcard(b, "catalog.product.created", func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	})
}

func TestBuilder_customRetryOptions(t *testing.T) {
	log := newFakeLogger()
	b := NewBuilder(log)
	b.WithRetryOptions(
		retry.Attempts(1),
		retry.Delay(1 * time.Millisecond),
	)
	Register(b, "catalog.product.created", func(ctx context.Context, ev testProductCreatedEvent) error {
		return errors.New("once")
	})

	d, err := b.Build()
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	// With 1 attempt, should return immediately
	msg := makeDelivery("catalog.product.created", testProductCreatedEvent{ProductID: "p1"})
	start := time.Now()
	err = d.Handle(context.Background(), msg)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected error")
	}
	// Should not have slept long (retry delay is 1ms)
	if elapsed > 50*time.Millisecond {
		t.Errorf("elapsed time %v suggests retries were applied unexpectedly", elapsed)
	}
}

// ---- validateKey tests ----

func TestValidateKey_emptyReturnsError(t *testing.T) {
	err := validateKey("")
	if err == nil {
		t.Error("expected error for empty key")
	}
}

func TestValidateKey_noDotReturnsError(t *testing.T) {
	err := validateKey("nokey")
	if err == nil {
		t.Error("expected error for key without dot")
	}
}

// ---- typedHandler tests ----

func TestTypedHandler_passesPayloadToHandler(t *testing.T) {
	var got testProductCreatedEvent
	h := NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		got = ev
		return nil
	})

	body := buildEnvelope("catalog.product.created", testProductCreatedEvent{ProductID: "p1", Title: "Bike"})
	raw := &amqp.Delivery{Body: body, RoutingKey: "catalog.product.created"}
	msg := types.NewDeliveryMessage(raw)

	if err := h.Handle(context.Background(), msg); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	if got.ProductID != "p1" || got.Title != "Bike" {
		t.Errorf("unexpected event: %+v", got)
	}
}

func TestTypedHandler_decodeError(t *testing.T) {
	h := NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	})

	raw := &amqp.Delivery{
		Body:       []byte(`{not json`),
		RoutingKey: "catalog.product.created",
	}
	msg := types.NewDeliveryMessage(raw)

	err := h.Handle(context.Background(), msg)
	if err == nil {
		t.Fatal("expected error from decode failure")
	}
}

func TestTypedHandler_wrongPayloadType(t *testing.T) {
	h := NewTypedHandler(func(ctx context.Context, ev testProductCreatedEvent) error {
		return nil
	})

	// Registered for testProductCreatedEvent but sending testProductDeletedEvent
	body := buildEnvelope("catalog.product.deleted", testProductDeletedEvent{ProductID: "p1"})
	raw := &amqp.Delivery{Body: body, RoutingKey: "catalog.product.deleted"}
	msg := types.NewDeliveryMessage(raw)

	err := h.Handle(context.Background(), msg)
	// json.Unmarshal into testProductCreatedEvent will succeed (both are objects)
	// but the fields won't match — this is expected behavior
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
}

// ---- Unused imports helper ----
var _ logger.Logger = (*fakeLogger)(nil) // assert fakeLogger satisfies logger.Logger
