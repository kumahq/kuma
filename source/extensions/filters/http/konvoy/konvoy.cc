#include "extensions/filters/http/konvoy/konvoy.h"

#include <string>

#include "extensions/filters/http/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

FilterConfig::FilterConfig(
    const envoy::config::filter::http::konvoy::v2alpha::Konvoy&,
    const LocalInfo::LocalInfo& local_info, Stats::Scope& scope,
    Runtime::Loader& runtime, Http::Context& http_context)
    : local_info_(local_info), scope_(scope),
      runtime_(runtime), http_context_(http_context) {}

Filter::Filter(FilterConfigSharedPtr config, Grpc::AsyncClientPtr&& async_client)
    : config_(config), async_client_(std::move(async_client)) {}

Filter::~Filter() {}

void Filter::onDestroy() {
  if (state_ == State::Calling) {
    state_ = State::Complete;
    stream_->resetStream();
    stream_ = nullptr;
  }
}

void Filter::setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) {
  decoder_callbacks_ = &callbacks;
}

Http::FilterHeadersStatus Filter::decodeHeaders(Http::HeaderMap& headers, bool end_stream) {
  ENVOY_LOG(info, "konvoy-filter: forwarding request headers to Konvoy (side car):\n{}", headers);

  state_ = State::Calling;

  stream_ = async_client_->start(
          *Protobuf::DescriptorPool::generated_pool()->FindMethodByName("envoy.service.konvoy.v2alpha.Konvoy.ProxyHttpRequest"), *this);

  auto message = KonvoyProtoUtils::requestHeadersMessage(headers);

  stream_->sendMessage(message, false);

  if (end_stream) {
    stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(), true);
  }

  return !end_stream ? Http::FilterHeadersStatus::Continue : Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::decodeData(Buffer::Instance& data, bool end_stream) {
  ENVOY_LOG(info, "konvoy-filter: forwarding request body to Konvoy (side car):\n{} bytes, end_stream={}", data.length(), end_stream);

  auto message = KonvoyProtoUtils::requestBodyChunckMessage(data);

  if (0 < data.length()) {
    // apparently, Envoy makes the last call to `decodeData` with an empty buffer and `end_stream` flag set
    stream_->sendMessage(message, false);
  }

  if (end_stream) {
    stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(), true);
  }

  return !end_stream ? Http::FilterDataStatus::Continue : Http::FilterDataStatus::StopIterationNoBuffer;
}

Http::FilterTrailersStatus Filter::decodeTrailers(Http::HeaderMap& trailers) {
  ENVOY_LOG(info, "konvoy-filter: forwarding request trailers to Konvoy (side car):\n{}", trailers);

  stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(trailers), true);

  return Http::FilterTrailersStatus::StopIteration;
}

/**
 * Called at the end of the stream, when all data has been decoded.
 */
void Filter::decodeComplete() {
  ENVOY_LOG(info, "konvoy-filter: forwarding is finished");
}

void Filter::onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart>&& message) {
  ENVOY_LOG(info, "konvoy-filter: received message from Konvoy (side car):\n{}", message->part_case());

  // TODO: pass data to the next Envoy filter
}

void Filter::onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) {
  ENVOY_LOG(info, "konvoy-filter: received close signal from Konvoy (side car):\nstatus = {}, message = {}", status, message);

  state_ = State::Complete;
  decoder_callbacks_->continueDecoding();
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
