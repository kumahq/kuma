#include "extensions/filters/http/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage KonvoyProtoUtils::requestHeadersMessage(
    const Http::HeaderMap& headers) {
    envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage message;
    message.mutable_request_headers();

    headers.iterate(
            [](const Http::HeaderEntry& header, void* context) -> Http::HeaderMap::Iterate {
                auto proto_header = static_cast<envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage*>(context)
                        ->mutable_request_headers()->mutable_headers()->add_headers();
                proto_header->set_key(header.key().c_str());
                proto_header->set_value(header.value().c_str());
                return Http::HeaderMap::Iterate::Continue;
            },
            &message);

    return message;
}

envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage KonvoyProtoUtils::requestBodyChunckMessage(
    const Buffer::Instance& data) {
    envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage message;
    message.mutable_request_body_chunk();

    message.mutable_request_body_chunk()->set_bytes(data.toString());

    return message;
}

envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage KonvoyProtoUtils::requestTrailersMessage() {
    envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage message;
    message.mutable_request_trailers();

    return message;
}

envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage KonvoyProtoUtils::requestTrailersMessage(
    const Http::HeaderMap& trailers) {
    envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage message;
    message.mutable_request_trailers();

    trailers.iterate(
            [](const Http::HeaderEntry& header, void* context) -> Http::HeaderMap::Iterate {
                auto proto_header = static_cast<envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage*>(context)
                        ->mutable_request_trailers()->mutable_trailers()->add_headers();
                proto_header->set_key(header.key().c_str());
                proto_header->set_value(header.value().c_str());
                return Http::HeaderMap::Iterate::Continue;
            },
            &message);

    return message;
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
