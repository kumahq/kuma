#pragma once

#include <string>

#include "envoy/grpc/async_client.h"
#include "envoy/server/filter_config.h"

#include "api/envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.h"
#include "api/envoy/service/konvoy/v2alpha/konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

class FilterConfig {
public:
  FilterConfig(const envoy::config::filter::http::konvoy::v2alpha::Konvoy &proto_config,
               const LocalInfo::LocalInfo& local_info, Stats::Scope& scope,
               Runtime::Loader& runtime, Http::Context& http_context);

  const LocalInfo::LocalInfo& localInfo() const { return local_info_; }
  Runtime::Loader& runtime() { return runtime_; }
  Stats::Scope& scope() { return scope_; }
  Http::Context& httpContext() { return http_context_; }

private:
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
  const FilterConfigSharedPtr config_;
  Grpc::AsyncClientPtr async_client_;

  Http::StreamDecoderFilterCallbacks *decoder_callbacks_;

  // State of this filter's communication with the external Konvoy service.
  // The filter has either not started calling the external service, in the middle of calling
  // it or has completed.
  enum class State { NotStarted, Calling, Complete };

  State state_{State::NotStarted};
  Grpc::AsyncStream* stream_{};
};

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
