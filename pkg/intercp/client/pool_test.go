package client_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/kumahq/kuma/pkg/intercp/client"
)

var _ = Describe("Pool", func() {
	It("should not create a client when TLS is not configured", func() {
		// given
		pool := client.NewPool(nil, 1*time.Second)

		// when
		_, err := pool.Client("http://192.168.0.1")

		// then
		Expect(err).To(Equal(client.TLSNotConfigured))
	})

	Context("configured with TLS", func() {
		var pool *client.Pool
		const idleDeadline = 100 * time.Millisecond

		BeforeEach(func() {
			pool = client.NewPool(func(s string, config *client.TLSConfig) (client.Conn, error) {
				return &testConn{
					state: connectivity.Ready,
				}, nil
			}, idleDeadline)
			pool.SetTLSConfig(&client.TLSConfig{})
			go pool.StartCleanup(context.Background(), 10*time.Millisecond)
		})

		It("should keep the connection open", func() {
			// when
			c, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())
			c2, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(c).To(BeIdenticalTo(c2))
		})

		It("should create a new connection after the previous expired", func() {
			// given
			c, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())

			// when
			time.Sleep(idleDeadline * 2)
			c2, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(c).NotTo(BeIdenticalTo(c2))
			Expect(c.(*testConn).closed).To(BeTrue())
		})

		It("should close broken connection and create a new one", func() {
			// given
			c, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())
			c.(*testConn).state = connectivity.TransientFailure

			// when
			c2, err := pool.Client("http://192.168.0.1")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(c).NotTo(BeIdenticalTo(c2))
			Expect(c.(*testConn).closed).To(BeTrue())
		})
	})
})

type testConn struct {
	*grpc.ClientConn
	closed bool
	state  connectivity.State
}

var _ client.Conn = &testConn{}

func (t *testConn) Close() error {
	t.closed = true
	return nil
}

func (t *testConn) GetState() connectivity.State {
	return t.state
}
