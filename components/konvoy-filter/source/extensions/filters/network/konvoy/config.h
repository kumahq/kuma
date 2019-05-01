#pragma once

#include "api/envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.h"
#include "api/envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.validate.h"

#include "extensions/filters/network/common/factory_base.h"
#include "extensions/filters/network/well_known_names.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

/**
 * Config registration for the Konvoy filter. @see NamedHttpFilterConfigFactory.
 */
class KonvoyFilterConfigFactory : public Common::FactoryBase<envoy::config::filter::network::konvoy::v2alpha::Konvoy> {
public:
    KonvoyFilterConfigFactory() : FactoryBase("konvoy") {}

private:
    Network::FilterFactoryCb createFilterFactoryFromProtoTyped(
            const envoy::config::filter::network::konvoy::v2alpha::Konvoy& proto_config,
            Server::Configuration::FactoryContext& context) override;
};

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
