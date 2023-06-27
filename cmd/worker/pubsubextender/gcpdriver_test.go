package pubsubextender

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	api "cloud.google.com/go/pubsub/apiv1"
	pb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"gocloud.dev/pubsub/gcppubsub"
	"gocloud.dev/pubsub/mempubsub"
	"golang.org/x/exp/slices"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	fakeServerPath = "projects/test/subscriptions/sub"
	ackDeadline    = 123
	fakeAckID      = "ackid-001"
)

type fakeSubServer struct {
	pb.UnimplementedSubscriberServer

	path            string
	subscription    *pb.Subscription
	lastAckDeadline int32
	lastAckIDs      []string
}

func newServer() *fakeSubServer {
	return &fakeSubServer{
		path: fakeServerPath,
		subscription: &pb.Subscription{
			AckDeadlineSeconds: ackDeadline,
		},
	}
}

func (f *fakeSubServer) GetSubscription(ctx context.Context, req *pb.GetSubscriptionRequest) (*pb.Subscription, error) {
	if f.path != req.Subscription {
		return nil, fmt.Errorf("unknown subscription: %s", req.Subscription)
	}
	return f.subscription, nil
}

func (f *fakeSubServer) ModifyAckDeadline(ctx context.Context, req *pb.ModifyAckDeadlineRequest) (*emptypb.Empty, error) {
	f.lastAckDeadline = req.GetAckDeadlineSeconds()
	f.lastAckIDs = req.GetAckIds()
	return &emptypb.Empty{}, nil
}

func (f *fakeSubServer) Pull(ctx context.Context, req *pb.PullRequest) (*pb.PullResponse, error) {
	if f.path != req.Subscription {
		return nil, fmt.Errorf("unknown subscription: %s", req.Subscription)
	}
	return &pb.PullResponse{
		ReceivedMessages: []*pb.ReceivedMessage{
			{
				AckId: fakeAckID,
				Message: &pb.PubsubMessage{
					Data:      []byte("Hello, world!"),
					MessageId: "msg-001",
				},
				DeliveryAttempt: 1,
			},
		},
	}, nil
}

func initTestClient(t *testing.T, ctx context.Context, server *fakeSubServer) (*api.SubscriberClient, func()) {
	t.Helper()
	lis := bufconn.Listen(4096 /* initial pipe buffer capacity */)
	fakeServerAddr := lis.Addr().String()

	gsrv := grpc.NewServer()
	pb.RegisterSubscriberServer(gsrv, server)
	go func() {
		if err := gsrv.Serve(lis); err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.DialContext(ctx, fakeServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}))
	if err != nil {
		panic(err)
	}

	// Create a client.
	ctxTimed, cancel := context.WithTimeout(ctx, 5*time.Second)
	client, err := api.NewSubscriberClient(ctxTimed, option.WithGRPCConn(conn))
	cancel()
	if err != nil {
		t.Fatal(err)
	}

	closer := func() {
		client.Close()
		err := lis.Close()
		if err != nil {
			panic(err)
		}
		gsrv.Stop()
	}

	return client, closer
}

func TestGCPNew(t *testing.T) {
	ctx := context.Background()
	client, closer := initTestClient(t, ctx, newServer())
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)
	e, err := New(ctx, "gcppubsub://"+fakeServerPath, sub)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	got := e.Deadline
	want := 123 * time.Second
	if got != want {
		t.Errorf("Deadline = %v; want %v", got, want)
	}
}

func TestNewGCPDriver_WrongScheme(t *testing.T) {
	u, err := url.Parse("kafka://a/b/c")
	if err != nil {
		t.Fatalf("Parse() = %v; want no error", err)
	}
	ctx := context.Background()
	client, closer := initTestClient(t, ctx, newServer())
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)

	d, err := newGCPDriver(u, sub)
	if err == nil {
		t.Errorf("newGCPDriver() = %v; want an error", err)
	}
	if d != nil {
		t.Errorf("newGCPDriver() = %v; want nil", d)
	}
}

func TestGCPNew_HostPath(t *testing.T) {
	ctx := context.Background()
	client, closer := initTestClient(t, ctx, newServer())
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)
	e, err := New(ctx, "gcppubsub://test/sub", sub)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	got := e.Deadline
	want := 123 * time.Second
	if got != want {
		t.Errorf("Deadline = %v; want %v", got, want)
	}
}

func TestNewGCPDriver_WrongSubscriptionDriver(t *testing.T) {
	ctx := context.Background()
	sub := mempubsub.NewSubscription(mempubsub.NewTopic(), 10*time.Second)
	e, err := New(ctx, "gcppubsub://test/sub", sub)
	if err == nil {
		t.Errorf("New() = %v; want an error", err)
	}
	if e != nil {
		t.Errorf("New() = %v; want nil", e)
	}
}

func TestGCPExtendAckDeadline(t *testing.T) {
	u, err := url.Parse("gcppubsub://" + fakeServerPath)
	if err != nil {
		t.Fatalf("Parse() = %v; want no error", err)
	}
	ctx := context.Background()
	srv := newServer()
	client, closer := initTestClient(t, ctx, srv)
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)
	d, err := newGCPDriver(u, sub)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	msg, err := sub.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive() = %v; want no error", err)
	}
	var wantDeadline int32 = 345
	wantAckIDs := []string{fakeAckID}
	err = d.ExtendMessageDeadline(ctx, msg, time.Duration(wantDeadline)*time.Second)
	if err != nil {
		t.Fatalf("ExtendMessageDeadline() = %v; want no error", err)
	}
	if got := srv.lastAckDeadline; got != wantDeadline {
		t.Errorf("Ack Deadline = %v; want %v", got, wantDeadline)
	}
	if got := srv.lastAckIDs; !slices.Equal(got, wantAckIDs) {
		t.Errorf("AckIDs = %v; want %v", got, wantAckIDs)
	}
}

func TestGCPExtendAckDeadline_LowerBound(t *testing.T) {
	u, err := url.Parse("gcppubsub://" + fakeServerPath)
	if err != nil {
		t.Fatalf("Parse() = %v; want no error", err)
	}
	ctx := context.Background()
	srv := newServer()
	client, closer := initTestClient(t, ctx, srv)
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)
	d, err := newGCPDriver(u, sub)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	msg, err := sub.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive() = %v; want no error", err)
	}
	var wantDeadline int32 = 10
	err = d.ExtendMessageDeadline(ctx, msg, 5*time.Second)
	if err != nil {
		t.Fatalf("ExtendMessageDeadline() = %v; want no error", err)
	}
	if got := srv.lastAckDeadline; got != wantDeadline {
		t.Errorf("Ack Deadline = %v; want %v", got, wantDeadline)
	}
}

func TestGCPExtendAckDeadline_UpperBound(t *testing.T) {
	u, err := url.Parse("gcppubsub://" + fakeServerPath)
	if err != nil {
		t.Fatalf("Parse() = %v; want no error", err)
	}
	ctx := context.Background()
	srv := newServer()
	client, closer := initTestClient(t, ctx, srv)
	defer closer()
	sub := gcppubsub.OpenSubscription(client, "test", "sub", nil)
	d, err := newGCPDriver(u, sub)
	if err != nil {
		t.Fatalf("New() = %v; want no error", err)
	}
	msg, err := sub.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive() = %v; want no error", err)
	}
	var wantDeadline int32 = 600
	err = d.ExtendMessageDeadline(ctx, msg, 1000*time.Second)
	if err != nil {
		t.Fatalf("ExtendMessageDeadline() = %v; want no error", err)
	}
	if got := srv.lastAckDeadline; got != wantDeadline {
		t.Errorf("Ack Deadline = %v; want %v", got, wantDeadline)
	}
}
