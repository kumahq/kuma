#pragma once

#include "envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.h"
#include "envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.validate.h"

#include "extensions/filters/http/common/factory_base.h"
#include "extensions/filters/http/well_known_names.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

/**
 * Config registration for the Konvoy filter. @see NamedHttpFilterConfigFactory.
 */
class KonvoyFilterConfigFactory : public Common::FactoryBase<envoy::config::filter::http::konvoy::v2alpha::Konvoy,
                                                      envoy::config::filter::http::konvoy::v2alpha::Konvoy> {
public:
    KonvoyFilterConfigFactory() : FactoryBase("konvoy") {}

private:
    Http::FilterFactoryCb createFilterFactoryFromProtoTyped(
            const envoy::config::filter::http::konvoy::v2alpha::Konvoy& proto_config,
            const std::string& stats_prefix, Server::Configuration::FactoryContext& context) override;
};

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
