#pragma once

#include <string>

#include "envoy/grpc/async_client.h"
#include "envoy/server/filter_config.h"

#include "envoy/stats/scope.h"
#include "envoy/stats/stats_macros.h"

#include "api/envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.h"
#include "api/envoy/service/konvoy/v2alpha/konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

/**
 * All Konvoy stats. @see stats_macros.h
 */
// clang-format off
#define ALL_HTTP_KONVOY_STATS(COUNTER, HISTOGRAM)     \
  COUNTER  (request_total)                            \
  COUNTER  (request_total_stream_latency_ms)          \
  COUNTER  (request_total_stream_start_latency_ms)    \
  COUNTER  (request_total_stream_exchange_latency_ms) \
  HISTOGRAM(request_stream_start_latency_ms)          \
  HISTOGRAM(request_stream_exchange_latency_ms)       \
  HISTOGRAM(request_stream_latency_ms)
// clang-format on

/**
 * Struct definition for all Konvoy stats. @see stats_macros.h
 */
struct InstanceStats {
    ALL_HTTP_KONVOY_STATS(GENERATE_COUNTER_STRUCT, GENERATE_HISTOGRAM_STRUCT)
};

class FilterConfig {
public:
  FilterConfig(const envoy::config::filter::http::konvoy::v2alpha::Konvoy &proto_config,
               const LocalInfo::LocalInfo& local_info, Stats::Scope& scope,
               Runtime::Loader& runtime, Http::Context& http_context, TimeSource& time_source);

  const InstanceStats& stats() { return stats_; }
  TimeSource& timeSource() const { return time_source_; }

  const LocalInfo::LocalInfo& localInfo() const { return local_info_; }
  Runtime::Loader& runtime() { return runtime_; }
  Stats::Scope& scope() { return scope_; }
  Http::Context& httpContext() { return http_context_; }

private:
  static InstanceStats generateStats(const std::string& name, Stats::Scope& scope);
  const InstanceStats stats_;
  TimeSource& time_source_;

  const LocalInfo::LocalInfo& local_info_;
  Stats::Scope& scope_;
  Runtime::Loader& runtime_;
  Http::Context& http_context_;
};

typedef std::shared_ptr<FilterConfig> FilterConfigSharedPtr;

typedef Grpc::TypedAsyncStreamCallbacks<envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart>
        KonvoyAsyncStreamCallbacks;

class Filter : public Logger::Loggable<Logger::Id::filter>,
               public Http::StreamDecoderFilter,
               public KonvoyAsyncStreamCallbacks {
public:
  Filter(FilterConfigSharedPtr, Grpc::AsyncClientPtr&& async_client);
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
  void onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart>&& message) override;
  void onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) override;

private:
  void endStreamIfNecessary(bool end_stream);
  void endStream(Http::HeaderMap& trailers);
  void chargeStreamStats(Grpc::Status::GrpcStatus status);

  const FilterConfigSharedPtr config_;
  Grpc::AsyncClientPtr async_client_;

  MonotonicTime start_stream_;
  MonotonicTime start_stream_complete_;
  Http::StreamDecoderFilterCallbacks *decoder_callbacks_;
  Http::HeaderMap* request_headers_;
  Http::HeaderMap* request_trailers_;

  // State of this filter's communication with the external Konvoy service.
  // The filter has either not started calling the external service, in the middle of calling
  // it or has completed.
  enum class State { NotStarted, Calling, Complete };

  State state_{State::NotStarted};

  const Protobuf::MethodDescriptor& service_method_;
  Grpc::AsyncStream* stream_{};
};

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
