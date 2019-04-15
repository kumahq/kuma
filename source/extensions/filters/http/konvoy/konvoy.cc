#include "extensions/filters/http/konvoy/konvoy.h"

#include <string>

#include "common/buffer/buffer_impl.h"
#include "common/common/enum_to_int.h"
#include "common/http/header_map_impl.h"
#include "common/http/header_utility.h"
#include "common/http/utility.h"

#include "extensions/filters/http/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

InstanceStats Config::generateStats(const std::string& name, Stats::Scope& scope) {
  const std::string final_prefix = fmt::format("konvoy.http.{}.", name);
  return {ALL_HTTP_KONVOY_STATS(
    POOL_GAUGE_PREFIX(scope, final_prefix),
    POOL_COUNTER_PREFIX(scope, final_prefix),
    POOL_HISTOGRAM_PREFIX(scope, final_prefix))};
}

Config::Config(
    const envoy::config::filter::http::konvoy::v2alpha::Konvoy& config,
    Stats::Scope& scope, TimeSource& time_source)
    : proto_config_(config), stats_(generateStats(config.stat_prefix(), scope)), time_source_(time_source),
      scope_(scope) {}

Filter::Filter(ConfigSharedPtr config, Grpc::AsyncClientPtr&& async_client)
    : config_(config),
    async_client_(std::move(async_client)),
    service_method_(*Protobuf::DescriptorPool::generated_pool()->FindMethodByName("envoy.service.konvoy.v2alpha.HttpKonvoy.ProxyHttpRequest")) {}

Filter::~Filter() {}

void Filter::onDestroy() {
  if (state_ == State::Streaming) {
    state_ = State::Complete;
    stream_->resetStream();
    config_->stats().rq_active_.dec();
    config_->stats().rq_cancel_.inc();
    chargeStreamStats(Grpc::Status::GrpcStatus::Canceled);
  }
}

void Filter::setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) {
  decoder_callbacks_ = &callbacks;
}

Http::FilterHeadersStatus Filter::decodeHeaders(Http::HeaderMap& headers, bool end_stream) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request headers to HTTP Konvoy Service (side car):\n{}", headers);

  // keep original headers for later modification
  request_headers_ = &headers;

  state_ = State::Streaming;

  config_->stats().rq_total_.inc();

  start_stream_ = config_->timeSource().monotonicTime();

  // need to increment in advance to support a scenario where `onRemoteClose` is called back from `start`
  config_->stats().rq_active_.inc();

  stream_ = async_client_->start(service_method_, *this);

  if (stream_ == nullptr) {
    ENVOY_LOG_MISC(debug, "konvoy-http-filter: failed to start a new stream to the Http Konvoy Service (side car)");
    // error handling must already have happened inside `onRemoteClose`
    ASSERT(state_ == State::Responded);
    return Http::FilterHeadersStatus::StopIteration;
  }

  if (config_->getProtoConfig().per_service_config().has_http_konvoy()) {
    auto &http_konvoy_config = config_->getProtoConfig().per_service_config().http_konvoy();

    auto config_message = KonvoyProtoUtils::serviceConfigurationMessage(http_konvoy_config);

    stream_->sendMessage(config_message, false);
  }

  auto message = KonvoyProtoUtils::requestHeadersMessage(headers);

  stream_->sendMessage(message, false);

  if (end_stream) {
    endStream(nullptr);
  }

  // don't pass request headers to the next filter yet
  return Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::decodeData(Buffer::Instance& data, bool end_stream) {
  ASSERT(state_ != State::Responded);

  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request body to HTTP Konvoy Service (side car):\n{} bytes, end_stream={}, buffer_size={}",
          data.length(), end_stream, decoder_callbacks_->decodingBuffer() ? decoder_callbacks_->decodingBuffer()->length() : 0);

  if (0 < data.length()) {
    // apparently, Envoy makes the last call to `decodeData` with an empty buffer and `end_stream` flag set

    auto message = KonvoyProtoUtils::requestBodyChunckMessage(data);

    stream_->sendMessage(message, false);
  }

  if (end_stream) {
    endStream(&decoder_callbacks_->addDecodedTrailers());
  }

  // don't pass request body to the next filter yet and don't buffer in the meantime
  return Http::FilterDataStatus::StopIterationNoBuffer;
}

Http::FilterTrailersStatus Filter::decodeTrailers(Http::HeaderMap& trailers) {
  ASSERT(state_ != State::Responded);

  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding request trailers to HTTP Konvoy Service (side car):\n{}", trailers);

  endStream(&trailers);

  // don't pass request trailers to the next filter yet
  return Http::FilterTrailersStatus::StopIteration;
}

/**
 * Called at the end of the stream, when all data has been decoded.
 */
void Filter::decodeComplete() {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: forwarding is finished");
}

void Filter::onReceiveMessage(std::unique_ptr <envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage> &&message) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: received message from HTTP Konvoy Service (side car):\n{}",
                 message->message_case());

  switch (message->message_case()) {
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kRequestHeaders: {

      // clear original headers
      request_headers_->removePrefix(Http::LowerCaseString{""});

      // add headers from the response
      auto headers = message->request_headers().headers();
      for (int index = 0; index < headers.headers_size(); index++) {
        auto &header = headers.headers(index);

        auto header_to_modify = request_headers_->get(Http::LowerCaseString(header.key()));
        if (header_to_modify) {
          header_to_modify->value(header.value().c_str(), header.value().size());
        } else {
          request_headers_->addCopy(Http::LowerCaseString(header.key()), header.value());
        }
      }

      break;
    }
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kRequestBodyChunk: {

      if (!message->request_body_chunk().bytes().empty()) {
        Buffer::OwnedImpl data{message->request_body_chunk().bytes()};

        decoder_callbacks_->addDecodedData(data, false);
      }

      break;
    }
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kRequestTrailers: {

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
      } else if (message->request_trailers().has_trailers()) {
        ENVOY_LOG_MISC(warn, "konvoy-http-filter: trailers from HTTP Konvoy Service will be ignored");
      }

      break;
    }
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kResponseHeaders: {

      if (!response_headers_) {
        response_headers_ = std::make_unique<Http::HeaderMapImpl>();
      }

      // add headers from the response
      auto headers = message->response_headers().headers();
      for (int index = 0; index < headers.headers_size(); index++) {
        auto &header = headers.headers(index);

        auto header_to_modify = response_headers_->get(Http::LowerCaseString(header.key()));
        if (header_to_modify) {
          header_to_modify->value(header.value().c_str(), header.value().size());
        } else {
          response_headers_->addCopy(Http::LowerCaseString(header.key()), header.value());
        }
      }

      break;
    }
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kResponseBodyChunk: {

      // TODO(yskopets): support response body

      break;
    }
    case envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage::MessageCase::kResponseTrailers: {

      uint64_t status_code = enumToInt(Http::Code::InternalServerError);

      const Http::HeaderEntry* status_header = response_headers_->Status();
      if (status_header) {
        status_code = Http::Utility::getResponseStatus(*response_headers_);
      }

      decoder_callbacks_->sendLocalReply(
              static_cast<Http::Code>(status_code),
              "",
              nullptr,
              absl::nullopt);

      break;
    }
    default:
      break;
  }
}

void Filter::onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) {
  ENVOY_LOG_MISC(trace, "konvoy-http-filter: received close signal from HTTP Konvoy Service (side car):\nstatus = {}, message = {}", status, message);

  state_ = State::Complete;
  config_->stats().rq_active_.dec();

  chargeStreamStats(status);

  if (status == Grpc::Status::GrpcStatus::Ok) {
    decoder_callbacks_->continueDecoding();
  } else {
    state_ = State::Responded;
    config_->stats().rq_error_.inc();
    decoder_callbacks_->sendLocalReply(Http::Code::InternalServerError, "", nullptr, absl::nullopt);
  }
}

void Filter::endStream(Http::HeaderMap* trailers) {
  // keep original trailers for later modification
  request_trailers_ = trailers;

  auto message = request_trailers_ ? KonvoyProtoUtils::requestTrailersMessage(*request_trailers_) : KonvoyProtoUtils::requestTrailersMessage();

  stream_->sendMessage(message, true);
}

void Filter::chargeStreamStats(Grpc::Status::GrpcStatus) {
  if (!stream_) {
    // if a stream has never been opened, stats make no sense
    return;
  }

  auto now = config_->timeSource().monotonicTime();

  std::chrono::milliseconds totalLatency = std::chrono::duration_cast<std::chrono::milliseconds>(now - start_stream_);

  config_->stats().rq_stream_latency_ms_.recordValue(totalLatency.count());
  config_->stats().rq_total_stream_latency_ms_.add(totalLatency.count());
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
