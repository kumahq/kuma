#pragma once

#include <string>

#include "envoy/grpc/async_client.h"

#include "envoy/network/connection.h"
#include "envoy/network/filter.h"
#include "envoy/runtime/runtime.h"

#include "envoy/stats/scope.h"
#include "envoy/stats/stats_macros.h"

#include "api/envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.h"
#include "api/envoy/service/konvoy/v2alpha/network_konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

/**
 * All Konvoy stats. @see stats_macros.h
 */
// clang-format off
#define ALL_TCP_KONVOY_STATS(GAUGE, COUNTER, HISTOGRAM)  \
  GAUGE    (cx_active)                                   \
  COUNTER  (cx_total)                                    \
  COUNTER  (cx_error)                                    \
  COUNTER  (cx_total_stream_latency_ms)                  \
  COUNTER  (cx_total_stream_start_latency_ms)            \
  COUNTER  (cx_total_stream_exchange_latency_ms)         \
  HISTOGRAM(cx_stream_start_latency_ms)                  \
  HISTOGRAM(cx_stream_exchange_latency_ms)               \
  HISTOGRAM(cx_stream_latency_ms)
// clang-format on

/**
 * Struct definition for all Konvoy stats. @see stats_macros.h
 */
struct InstanceStats {
    ALL_TCP_KONVOY_STATS(GENERATE_GAUGE_STRUCT, GENERATE_COUNTER_STRUCT, GENERATE_HISTOGRAM_STRUCT)
};

/**
 * Configuration for TCP Konvoy filter.
 */
class Config {
public:
    Config(const envoy::config::filter::network::konvoy::v2alpha::Konvoy &proto_config,
                 Stats::Scope& scope, Runtime::Loader& runtime, TimeSource& time_source);

    const InstanceStats& stats() { return stats_; }
    TimeSource& timeSource() const { return time_source_; }

    Runtime::Loader& runtime() { return runtime_; }
    Stats::Scope& scope() { return scope_; }

private:
    static InstanceStats generateStats(const std::string& name, Stats::Scope& scope);
    const InstanceStats stats_;
    TimeSource& time_source_;

    Stats::Scope& scope_;
    Runtime::Loader& runtime_;
};

typedef std::shared_ptr<Config> ConfigSharedPtr;

typedef Grpc::TypedAsyncStreamCallbacks<envoy::service::konvoy::v2alpha::KonvoyProxyConnectionResponseMessage>
        NetworkKonvoyAsyncStreamCallbacks;

class Filter : public Logger::Loggable<Logger::Id::filter>,
               public Network::ReadFilter,
               public Network::ConnectionCallbacks,
               public NetworkKonvoyAsyncStreamCallbacks {
public:
    Filter(ConfigSharedPtr, Grpc::AsyncClientPtr&& async_client);
    ~Filter();

    // Network::ReadFilter
    Network::FilterStatus onData(Buffer::Instance& data, bool end_stream) override;
    Network::FilterStatus onNewConnection() override;
    void initializeReadFilterCallbacks(Network::ReadFilterCallbacks&) override;

    // Network::ConnectionCallbacks
    void onEvent(Network::ConnectionEvent event) override;
    void onAboveWriteBufferHighWatermark() override {}
    void onBelowWriteBufferLowWatermark() override {}

    // Grpc::AsyncStreamCallbacks
    void onCreateInitialMetadata(Http::HeaderMap&) override {}
    void onReceiveInitialMetadata(Http::HeaderMapPtr&&) override {}
    void onReceiveTrailingMetadata(Http::HeaderMapPtr&&) override {}
    void onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::KonvoyProxyConnectionResponseMessage>&& message) override;
    void onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) override;

private:
    void endStreamIfNecessary(bool end_stream);
    void endStream();
    void chargeStreamStats(Grpc::Status::GrpcStatus status);

    const ConfigSharedPtr config_;
    Grpc::AsyncClientPtr async_client_;

    MonotonicTime start_stream_;
    MonotonicTime start_stream_complete_;
    Network::ReadFilterCallbacks* read_callbacks_{};

    Buffer::Instance* buffer_;

    // State of this filter's communication with the external Konvoy service.
    // The filter has either not started calling the external service, in the middle of calling
    // it or has completed.
    enum class State { NotStarted, Calling, Complete, Responded };

    State state_{State::NotStarted};

    const Protobuf::MethodDescriptor& service_method_;
    Grpc::AsyncStream* stream_{};
};

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
