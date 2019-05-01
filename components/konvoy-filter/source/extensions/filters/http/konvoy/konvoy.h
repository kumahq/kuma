#pragma once

#include <string>

#include "envoy/grpc/async_client.h"
#include "envoy/server/filter_config.h"

#include "envoy/stats/scope.h"
#include "envoy/stats/stats_macros.h"

#include "envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.h"
#include "envoy/service/konvoy/v2alpha/http_konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

/**
 * All Konvoy stats. @see stats_macros.h
 */
// clang-format off
#define ALL_HTTP_KONVOY_STATS(GAUGE, COUNTER, HISTOGRAM) \
  GAUGE    (rq_active)                                   \
  COUNTER  (rq_total)                                    \
  COUNTER  (rq_error)                                    \
  COUNTER  (rq_cancel)                                   \
  COUNTER  (rq_total_stream_latency_ms)                  \
  HISTOGRAM(rq_stream_latency_ms)
// clang-format on

/**
 * Struct definition for all Konvoy stats. @see stats_macros.h
 */
struct InstanceStats {
    ALL_HTTP_KONVOY_STATS(GENERATE_GAUGE_STRUCT, GENERATE_COUNTER_STRUCT, GENERATE_HISTOGRAM_STRUCT)
};

class Config {
public:
  Config(const envoy::config::filter::http::konvoy::v2alpha::Konvoy &proto_config,
               Stats::Scope& scope, TimeSource& time_source);

  const envoy::config::filter::http::konvoy::v2alpha::Konvoy& getProtoConfig() const { return proto_config_; }
  const InstanceStats& stats() { return stats_; }
  TimeSource& timeSource() const { return time_source_; }

  Stats::Scope& scope() { return scope_; }

private:
  static InstanceStats generateStats(const std::string& name, Stats::Scope& scope);

  envoy::config::filter::http::konvoy::v2alpha::Konvoy proto_config_;
  const InstanceStats stats_;
  TimeSource& time_source_;

  Stats::Scope& scope_;
};

typedef std::shared_ptr<Config> ConfigSharedPtr;

typedef Grpc::TypedAsyncStreamCallbacks<envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage>
        HttpKonvoyAsyncStreamCallbacks;

class Filter : public Logger::Loggable<Logger::Id::filter>,
               public Http::StreamDecoderFilter,
               public HttpKonvoyAsyncStreamCallbacks {
public:
  Filter(ConfigSharedPtr, Grpc::AsyncClientPtr&& async_client);
  ~Filter();

  void setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks &) override;

  // Http::StreamFilterBase
  void onDestroy() override;

  // Http::StreamDecoderFilter
  Http::FilterHeadersStatus decodeHeaders(Http::HeaderMap &, bool) override;
  Http::FilterDataStatus decodeData(Buffer::Instance &, bool) override;
  Http::FilterTrailersStatus decodeTrailers(Http::HeaderMap &) override;
  void decodeComplete() override;

  // Grpc::AsyncStreamCallbacks
  void onCreateInitialMetadata(Http::HeaderMap&) override {}
  void onReceiveInitialMetadata(Http::HeaderMapPtr&&) override {}
  void onReceiveTrailingMetadata(Http::HeaderMapPtr&&) override {}
  void onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage>&& message) override;
  void onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) override;

private:
  void endStream(Http::HeaderMap* trailers);
  void chargeStreamStats(Grpc::Status::GrpcStatus status);

  const ConfigSharedPtr config_;
  Grpc::AsyncClientPtr async_client_;

  MonotonicTime start_stream_;
  Http::StreamDecoderFilterCallbacks *decoder_callbacks_;
  Http::HeaderMap* request_headers_;
  Http::HeaderMap* request_trailers_;
  Http::HeaderMapPtr response_headers_;

  // State of this filter's communication with the HTTP Konvoy Service.
  // The filter has either not started streaming to the HTTP Konvoy Service,
  // in the middle of streaming or has completed.
  enum class State { NotStarted, Streaming, Complete, Responded };

  State state_{State::NotStarted};

  const Protobuf::MethodDescriptor& service_method_;
  Grpc::AsyncStream* stream_{};
};

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
