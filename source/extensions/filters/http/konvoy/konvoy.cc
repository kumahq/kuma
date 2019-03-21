#include "extensions/filters/http/konvoy/konvoy.h"

#include <string>

#include "common/buffer/buffer_impl.h"

#include "extensions/filters/http/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

InstanceStats FilterConfig::generateStats(const std::string& name, Stats::Scope& scope) {
  const std::string final_prefix = fmt::format("konvoy.http.{}.", name);
  return {ALL_HTTP_KONVOY_STATS(POOL_COUNTER_PREFIX(scope, final_prefix), POOL_HISTOGRAM_PREFIX(scope, final_prefix))};
}

FilterConfig::FilterConfig(
    const envoy::config::filter::http::konvoy::v2alpha::Konvoy& config,
    const LocalInfo::LocalInfo& local_info, Stats::Scope& scope,
    Runtime::Loader& runtime, Http::Context& http_context, TimeSource& time_source)
    : stats_(generateStats(config.stat_prefix(), scope)), time_source_(time_source),
      local_info_(local_info), scope_(scope),
      runtime_(runtime), http_context_(http_context) {}

Filter::Filter(FilterConfigSharedPtr config, Grpc::AsyncClientPtr&& async_client)
    : config_(config),
    async_client_(std::move(async_client)),
    service_method_(*Protobuf::DescriptorPool::generated_pool()->FindMethodByName("envoy.service.konvoy.v2alpha.HttpKonvoy.ProxyHttpRequest")) {}

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
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request headers to HTTP Konvoy Service (side car):\n{}", headers);

  // keep original headers for later modification
  request_headers_ = &headers;

  state_ = State::Calling;

  config_->stats().request_total_.inc();

  start_stream_ = config_->timeSource().monotonicTime();

  stream_ = async_client_->start(service_method_, *this);

  start_stream_complete_ = config_->timeSource().monotonicTime();

  auto message = KonvoyProtoUtils::requestHeadersMessage(headers);

  stream_->sendMessage(message, false);

  endStreamIfNecessary(end_stream);

  // don't pass request headers to the next filter yet
  return Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::decodeData(Buffer::Instance& data, bool end_stream) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request body to HTTP Konvoy Service (side car):\n{} bytes, end_stream={}, buffer_size={}",
          data.length(), end_stream, decoder_callbacks_->decodingBuffer() ? decoder_callbacks_->decodingBuffer()->length() : 0);

  if (0 < data.length()) {
    // apparently, Envoy makes the last call to `decodeData` with an empty buffer and `end_stream` flag set

    auto message = KonvoyProtoUtils::requestBodyChunckMessage(data);

    stream_->sendMessage(message, false);
  }

  endStreamIfNecessary(end_stream);

  // don't pass request body to the next filter yet and don't buffer in the meantime
  return Http::FilterDataStatus::StopIterationNoBuffer;
}

Http::FilterTrailersStatus Filter::decodeTrailers(Http::HeaderMap& trailers) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request trailers to HTTP Konvoy Service (side car):\n{}", trailers);

  endStream(trailers);

  // don't pass request trailers to the next filter yet
  return Http::FilterTrailersStatus::StopIteration;
}

/**
 * Called at the end of the stream, when all data has been decoded.
 */
void Filter::decodeComplete() {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding is finished");
}

void Filter::onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart>&& message) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: received message from HTTP Konvoy Service (side car):\n{}", message->part_case());

  switch (message->part_case()) {
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestHeaders: {

          // clear original headers
          request_headers_->removePrefix(Http::LowerCaseString{""});

          // add headers from the response
          auto headers = message->request_headers().headers();
          for (int index = 0; index < headers.headers_size(); index++) {
              auto& header = headers.headers(index);

              auto header_to_modify = request_headers_->get(Http::LowerCaseString(header.key()));
              if (header_to_modify) {
                  header_to_modify->value(header.value().c_str(), header.value().size());
              } else {
                  request_headers_->addCopy(Http::LowerCaseString(header.key()), header.value());
              }
          }

          break;
      }
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestBodyChunk: {

          if (!message->request_body_chunk().bytes().empty()) {
              Buffer::OwnedImpl data{message->request_body_chunk().bytes()};

              decoder_callbacks_->addDecodedData(data, false);
          }

          break;
      }
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestTrailers: {

          if (request_trailers_) {
              // clear original trailers
              request_trailers_->removePrefix(Http::LowerCaseString{""});

              // add trailers from the response
              auto trailers = message->request_trailers().trailers();
              for (int index = 0; index < trailers.headers_size(); index++) {
                  auto &trailer = trailers.headers(index);

                  auto trailer_to_modify = request_trailers_->get(Http::LowerCaseString(trailer.key()));
                  if (trailer_to_modify) {
                      trailer_to_modify->value(trailer.value().c_str(), trailer.value().size());
                  } else {
                      request_trailers_->addCopy(Http::LowerCaseString(trailer.key()), trailer.value());
                  }
              }
          }

          break;
      }
      default:
          break;
  }
}

void Filter::onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: received close signal from HTTP Konvoy Service (side car):\nstatus = {}, message = {}", status, message);

  state_ = State::Complete;

  chargeStreamStats(status);

  decoder_callbacks_->continueDecoding();
}

void Filter::endStreamIfNecessary(bool end_stream) {
  if (end_stream) {
    endStream(decoder_callbacks_->addDecodedTrailers());
  }
}

void Filter::endStream(Http::HeaderMap& trailers) {
  // keep original trailers for later modification
  request_trailers_ = &trailers;

  auto message = KonvoyProtoUtils::requestTrailersMessage(trailers);

  stream_->sendMessage(message, true);
}

void Filter::chargeStreamStats(Grpc::Status::GrpcStatus) {
  auto now = config_->timeSource().monotonicTime();

  std::chrono::milliseconds totalLatency = std::chrono::duration_cast<std::chrono::milliseconds>(now - start_stream_);

  config_->stats().request_stream_latency_ms_.recordValue(totalLatency.count());
  config_->stats().request_total_stream_latency_ms_.add(totalLatency.count());

  std::chrono::milliseconds startLatency = std::chrono::duration_cast<std::chrono::milliseconds>(start_stream_complete_ - start_stream_);

  config_->stats().request_stream_start_latency_ms_.recordValue(startLatency.count());
  config_->stats().request_total_stream_start_latency_ms_.add(startLatency.count());

  std::chrono::milliseconds exchangeLatency = std::chrono::duration_cast<std::chrono::milliseconds>(now - start_stream_complete_);

  config_->stats().request_stream_exchange_latency_ms_.recordValue(exchangeLatency.count());
  config_->stats().request_total_stream_exchange_latency_ms_.add(exchangeLatency.count());
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
