package client_test

import (
	std_errors "errors"
	"os"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	client_v2 "github.com/kumahq/kuma/v2/pkg/kds/v2/client"
)

const debugKDSPayloadDumpEnv = "KUMA_DEBUG_KDS_DUMP"

type recordedLogEntry struct {
	level         int
	name          string
	message       string
	keysAndValues []any
}

type recordingSink struct {
	entries *[]recordedLogEntry
	name    string
	values  []any
}

func (r *recordingSink) Init(logr.RuntimeInfo) {}

func (r *recordingSink) Enabled(int) bool {
	return true
}

func (r *recordingSink) Info(level int, msg string, keysAndValues ...any) {
	entryValues := append([]any{}, r.values...)
	entryValues = append(entryValues, keysAndValues...)
	*r.entries = append(*r.entries, recordedLogEntry{
		level:         level,
		name:          r.name,
		message:       msg,
		keysAndValues: entryValues,
	})
}

func (r *recordingSink) Error(error, string, ...any) {}

func (r *recordingSink) WithValues(keysAndValues ...any) logr.LogSink {
	nextValues := append([]any{}, r.values...)
	nextValues = append(nextValues, keysAndValues...)

	return &recordingSink{
		entries: r.entries,
		name:    r.name,
		values:  nextValues,
	}
}

func (r *recordingSink) WithName(name string) logr.LogSink {
	nextName := name
	if r.name != "" {
		nextName = r.name + "/" + name
	}

	return &recordingSink{
		entries: r.entries,
		name:    nextName,
		values:  append([]any{}, r.values...),
	}
}

type fakeDeltaKDSStream struct {
	responses []client_v2.UpstreamResponse
	ackErr    error
	nackErr   error
}

func (f *fakeDeltaKDSStream) DeltaDiscoveryRequest(core_model.ResourceType) error {
	return nil
}

func (f *fakeDeltaKDSStream) Receive() (client_v2.UpstreamResponse, error) {
	if len(f.responses) == 0 {
		return client_v2.UpstreamResponse{}, std_errors.New("unexpected receive call")
	}

	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeDeltaKDSStream) ACK(core_model.ResourceType) error {
	return f.ackErr
}

func (f *fakeDeltaKDSStream) NACK(core_model.ResourceType, error) error {
	return f.nackErr
}

func (f *fakeDeltaKDSStream) CloseSend() error {
	return nil
}

var _ = Describe("KDS sync client logging", func() {
	var entries []recordedLogEntry
	var stopErr error

	setDebugPayloadDump := func(value string) {
		originalValue, hadOriginal := os.LookupEnv(debugKDSPayloadDumpEnv)
		Expect(os.Setenv(debugKDSPayloadDumpEnv, value)).To(Succeed())
		DeferCleanup(func() {
			if hadOriginal {
				Expect(os.Setenv(debugKDSPayloadDumpEnv, originalValue)).To(Succeed())
				return
			}

			Expect(os.Unsetenv(debugKDSPayloadDumpEnv)).To(Succeed())
		})
	}

	newClient := func(stream *fakeDeltaKDSStream, callbacks *client_v2.Callbacks) client_v2.KDSSyncClient {
		return client_v2.NewKDSSyncClient(
			logr.New(&recordingSink{entries: &entries}),
			[]core_model.ResourceType{mesh.MeshType},
			stream,
			callbacks,
			0,
		)
	}

	BeforeEach(func() {
		entries = []recordedLogEntry{}
		stopErr = std_errors.New("stop")
	})

	It("should log a summary by default", func() {
		// setup
		setDebugPayloadDump("false")
		stream := &fakeDeltaKDSStream{
			responses: []client_v2.UpstreamResponse{
				{
					Type:                mesh.MeshType,
					Nonce:               "nonce-1",
					AddedResources:      &mesh.MeshResourceList{},
					RemovedResourcesKey: make([]core_model.ResourceKey, 1),
				},
			},
			ackErr: stopErr,
		}
		client := newClient(stream, nil)

		// when
		err := client.Receive()

		// then
		Expect(err).To(MatchError(ContainSubstring("failed to ACK a discovery response")))
		Expect(logFields(findLogEntry(entries, "DeltaDiscoveryResponse received"))).To(Equal(map[string]any{
			"type":                  mesh.MeshType,
			"nonce":                 "nonce-1",
			"addedResourcesCount":   0,
			"removedResourcesCount": 1,
		}))
		Expect(logFields(findLogEntry(entries, "no callback set, sending ACK"))).To(Equal(map[string]any{
			"type":  string(mesh.MeshType),
			"nonce": "nonce-1",
		}))
	})

	It("should log the full response when the debug dump is enabled", func() {
		// setup
		setDebugPayloadDump("true")
		response := client_v2.UpstreamResponse{
			Type:           mesh.MeshType,
			Nonce:          "nonce-2",
			AddedResources: &mesh.MeshResourceList{},
		}
		stream := &fakeDeltaKDSStream{
			responses: []client_v2.UpstreamResponse{response},
			ackErr:    stopErr,
		}
		client := newClient(stream, nil)

		// when
		err := client.Receive()

		// then
		Expect(err).To(MatchError(ContainSubstring("failed to ACK a discovery response")))
		Expect(logFields(findLogEntry(entries, "DeltaDiscoveryResponse received"))).To(Equal(map[string]any{
			"response": response,
		}))
	})

	It("should include the nonce on ACK logs with callbacks", func() {
		// setup
		stream := &fakeDeltaKDSStream{
			responses: []client_v2.UpstreamResponse{
				{
					Type:           mesh.MeshType,
					Nonce:          "nonce-ack",
					AddedResources: &mesh.MeshResourceList{},
				},
			},
			ackErr: stopErr,
		}
		client := newClient(stream, &client_v2.Callbacks{
			OnResourcesReceived: func(client_v2.UpstreamResponse) (error, error) {
				return nil, nil
			},
		})

		// when
		err := client.Receive()

		// then
		Expect(err).To(MatchError(ContainSubstring("failed to ACK a discovery response")))
		Expect(logFields(findLogEntry(entries, "sending ACK"))).To(Equal(map[string]any{
			"type":  mesh.MeshType,
			"nonce": "nonce-ack",
		}))
	})

	It("should include the nonce on NACK logs", func() {
		// setup
		stream := &fakeDeltaKDSStream{
			responses: []client_v2.UpstreamResponse{
				{
					Type:           mesh.MeshType,
					Nonce:          "nonce-nack",
					AddedResources: &mesh.MeshResourceList{},
				},
			},
			nackErr: stopErr,
		}
		client := newClient(stream, &client_v2.Callbacks{
			OnResourcesReceived: func(client_v2.UpstreamResponse) (error, error) {
				return nil, std_errors.New("nack")
			},
		})

		// when
		err := client.Receive()

		// then
		Expect(err).To(MatchError(ContainSubstring("failed to NACK a discovery response")))
		Expect(logFields(findLogEntry(entries, "received resource is invalid, sending NACK"))["nonce"]).To(Equal("nonce-nack"))
	})
})

func findLogEntry(entries []recordedLogEntry, message string) recordedLogEntry {
	GinkgoHelper()

	for _, entry := range entries {
		if entry.message == message {
			return entry
		}
	}

	Fail("log entry not found: " + message)
	return recordedLogEntry{}
}

func logFields(entry recordedLogEntry) map[string]any {
	fields := map[string]any{}

	for i := 0; i < len(entry.keysAndValues); i += 2 {
		key, ok := entry.keysAndValues[i].(string)
		if !ok {
			continue
		}

		fields[key] = entry.keysAndValues[i+1]
	}

	return fields
}
