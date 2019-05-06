#pragma once

#include "envoy/buffer/buffer.h"
#include "envoy/service/konvoy/v2alpha/network_konvoy_service.pb.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

class KonvoyProtoUtils {
public:
    static envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage
    serviceConfigurationMessage(const ::google::protobuf::Any& config);

    static envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage
    requestDataChunkMessage(const Buffer::Instance& data);
};

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
