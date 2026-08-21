package service

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("recvError", func() {
	// The zone tells an oversized message apart from a generic transport
	// failure by the status code, so ResourceExhausted has to survive the
	// trip through the global CP's stream handler.
	DescribeTable("maps a recv error to the status the rpc terminates with",
		func(given error, expected codes.Code) {
			Expect(status.Code(recvError(given))).To(Equal(expected))
		},
		Entry("message exceeds the receive limit",
			status.Error(codes.ResourceExhausted, "grpc: message after decompression larger than max (20000000 vs. 10485760)"),
			codes.ResourceExhausted,
		),
		Entry("transport failure", io.ErrUnexpectedEOF, codes.Internal),
		Entry("plain error", errors.New("boom"), codes.Internal),
		Entry("other status code", status.Error(codes.Unavailable, "unavailable"), codes.Internal),
	)
})
