#include "extensions/filters/http/konvoy/config.h"

#include "extensions/filters/http/konvoy/konvoy.h"

#include "api/envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.validate.h"

#include "envoy/registry/registry.h"

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

Http::FilterFactoryCb KonvoyFilterConfig::createFilterFactoryFromProtoTyped(
    const envoy::config::filter::http::konvoy::v2alpha::Konvoy& proto_config, const std::string&,
    Server::Configuration::FactoryContext& context) {
  const auto filter_config = std::make_shared<FilterConfig>(
      proto_config, context.localInfo(), context.scope(), context.runtime(), context.httpContext());
  Http::FilterFactoryCb callback;

  // gRPC client.
  callback = [grpc_service = proto_config.grpc_service(), &context, filter_config](Http::FilterChainFactoryCallbacks& callbacks) {
    const auto async_client_factory =
        context.clusterManager().grpcAsyncClientManager().factoryForGrpcService(
            grpc_service, context.scope(), true);

      auto client = async_client_factory->create();

      callbacks.addStreamDecoderFilter(Http::StreamDecoderFilterSharedPtr{
        std::make_shared<Filter>(filter_config, std::move(client))});
  };

  return callback;
};

Router::RouteSpecificFilterConfigConstSharedPtr
KonvoyFilterConfig::createRouteSpecificFilterConfigTyped(
    const envoy::config::filter::http::konvoy::v2alpha::Konvoy&,
    Server::Configuration::FactoryContext&) {
  return nullptr;
}

/**
 * Static registration for the Konvoy filter. @see RegisterFactory.
 */
REGISTER_FACTORY(KonvoyFilterConfig, Server::Configuration::NamedHttpFilterConfigFactory);

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
