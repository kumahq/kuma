#include "extensions/filters/http/konvoy/konvoy.h"

#include <string>

#include "common/buffer/buffer_impl.h"

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
  ENVOY_LOG_MISC(trace, "konvoy-filter: forwarding request headers to Konvoy (side car):\n{}", headers);

  // keep original headers for later modification
  request_headers_ = &headers;

  state_ = State::Calling;

  stream_ = async_client_->start(
          *Protobuf::DescriptorPool::generated_pool()->FindMethodByName("envoy.service.konvoy.v2alpha.Konvoy.ProxyHttpRequest"), *this);

  auto message = KonvoyProtoUtils::requestHeadersMessage(headers);

  stream_->sendMessage(message, false);

  if (end_stream) {
    stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(), true);
  }

  // don't pass request headers to the next filter yet
  return Http::FilterHeadersStatus::StopIteration;
}

Http::FilterDataStatus Filter::decodeData(Buffer::Instance& data, bool end_stream) {
  ENVOY_LOG_MISC(trace, "konvoy-filter: forwarding request body to Konvoy (side car):\n{} bytes, end_stream={}, buffer_size={}",
          data.length(), end_stream, decoder_callbacks_->decodingBuffer() ? decoder_callbacks_->decodingBuffer()->length() : 0);

  auto message = KonvoyProtoUtils::requestBodyChunckMessage(data);

  if (0 < data.length()) {
    // apparently, Envoy makes the last call to `decodeData` with an empty buffer and `end_stream` flag set
    stream_->sendMessage(message, false);
  }

  if (end_stream) {
    // keep original trailers for later modification
    request_trailers_ = &decoder_callbacks_->addDecodedTrailers();

    stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(), true);
  }

  // don't pass request body to the next filter yet and don't buffer in the meantime
  return Http::FilterDataStatus::StopIterationNoBuffer;
}

Http::FilterTrailersStatus Filter::decodeTrailers(Http::HeaderMap& trailers) {
  ENVOY_LOG_MISC(trace, "konvoy-filter: forwarding request trailers to Konvoy (side car):\n{}", trailers);

  // keep original trailers for later modification
  request_trailers_ = &trailers;

  stream_->sendMessage(KonvoyProtoUtils::requestTrailersMessage(trailers), true);

  // don't pass request trailers to the next filter yet
  return Http::FilterTrailersStatus::StopIteration;
}

/**
 * Called at the end of the stream, when all data has been decoded.
 */
void Filter::decodeComplete() {
  ENVOY_LOG_MISC(trace, "konvoy-filter: forwarding is finished");
}

void Filter::onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart>&& message) {
  ENVOY_LOG_MISC(trace, "konvoy-filter: received message from Konvoy (side car):\n{}", message->part_case());

  switch (message->part_case()) {
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestHeaders: {

          // clear original headers
          request_headers_->iterate(
                  [](const Http::HeaderEntry &header, void *context) -> Http::HeaderMap::Iterate {
                      auto headers = static_cast<Http::HeaderMap *>(context);
                      headers->remove(Http::LowerCaseString(header.key().c_str()));
                      return Http::HeaderMap::Iterate::Continue;
                  },
                  request_headers_);

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
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestBodyChunk:

          if (0 < message->request_body_chunk().bytes().size()) {
              Buffer::OwnedImpl data;
              data.add(message->request_body_chunk().bytes());

              decoder_callbacks_->addDecodedData(data, false);
          }

          break;
      case envoy::service::konvoy::v2alpha::KonvoyHttpResponsePart::PartCase::kRequestTrailers: {

          if (request_trailers_) {
              // clear original trailers
              request_trailers_->iterate(
                      [](const Http::HeaderEntry &header, void *context) -> Http::HeaderMap::Iterate {
                          auto headers = static_cast<Http::HeaderMap *>(context);
                          headers->remove(Http::LowerCaseString(header.key().c_str()));
                          return Http::HeaderMap::Iterate::Continue;
                      },
                      request_trailers_);

              // add trailers from the response
              auto headers = message->request_headers().headers();
              for (int index = 0; index < headers.headers_size(); index++) {
                  auto &header = headers.headers(index);

                  auto header_to_modify = request_trailers_->get(Http::LowerCaseString(header.key()));
                  if (header_to_modify) {
                      header_to_modify->value(header.value().c_str(), header.value().size());
                  } else {
                      request_trailers_->addCopy(Http::LowerCaseString(header.key()), header.value());
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
  ENVOY_LOG_MISC(trace, "konvoy-filter: received close signal from Konvoy (side car):\nstatus = {}, message = {}", status, message);

  state_ = State::Complete;

  decoder_callbacks_->continueDecoding();
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
