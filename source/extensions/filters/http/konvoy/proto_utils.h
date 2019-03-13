#pragma once

#include "envoy/buffer/buffer.h"
#include "envoy/http/header_map.h"
#include "api/envoy/service/konvoy/v2alpha/konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

class KonvoyProtoUtils {
public:
    static envoy::service::konvoy::v2alpha::KonvoyHttpRequestPart
    requestHeadersMessage(const Http::HeaderMap& headers);

    static envoy::service::konvoy::v2alpha::KonvoyHttpRequestPart
    requestBodyChunckMessage(const Buffer::Instance& data);

    static envoy::service::konvoy::v2alpha::KonvoyHttpRequestPart
    requestTrailersMessage();

    static envoy::service::konvoy::v2alpha::KonvoyHttpRequestPart
    requestTrailersMessage(const Http::HeaderMap& trailers);
};

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
