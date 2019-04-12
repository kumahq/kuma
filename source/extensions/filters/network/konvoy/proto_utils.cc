#include "extensions/filters/network/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage KonvoyProtoUtils::serviceConfigurationMessage(
        const ::google::protobuf::Any& config) {
    envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage message;

    message.mutable_configuration()->mutable_config()->CopyFrom(config);

    return message;
}

envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage KonvoyProtoUtils::requestDataChunkMessage(
    const Buffer::Instance& data) {
    envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage message;
    message.mutable_request_data_chunk();

    message.mutable_request_data_chunk()->set_bytes(data.toString());

    return message;
}

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
