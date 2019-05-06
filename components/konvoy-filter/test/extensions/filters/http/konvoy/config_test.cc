#include "envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.h"
#include "envoy/config/filter/http/konvoy/v2alpha/konvoy.pb.validate.h"

#include "envoy/stats/scope.h"

#include "extensions/filters/http/konvoy/config.h"
#include "extensions/filters/http/konvoy/konvoy.h"

#include "test/mocks/server/mocks.h"

#include "gmock/gmock.h"
#include "gtest/gtest.h"

using testing::_;
using testing::Invoke;
using testing::Pointee;
using testing::StrictMock;
using testing::WhenDynamicCastTo;

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

TEST(KonvoyFilterConfigFactoryTest, InvalidProto) {
  // given
  envoy::config::filter::http::konvoy::v2alpha::Konvoy proto_config;

  // where
  NiceMock<Server::Configuration::MockFactoryContext> context;
  KonvoyFilterConfigFactory factory;

  // expect
  EXPECT_THROW(factory.createFilterFactoryFromProto(
          proto_config, "stats", context), ProtoValidationException);
}

TEST(KonvoyFilterConfigFactoryTest, MinimalValidProto) {
  // given
  const auto yaml = R"EOF(
  stat_prefix: demo-grpc-server
  grpc_service:
    envoy_grpc:
      cluster_name: konvoy_demo_side_car
  )EOF";

  // where
  KonvoyFilterConfigFactory factory;
  auto proto_config = factory.createEmptyConfigProto();
  MessageUtil::loadFromYaml(yaml, *proto_config);

  // expect
  NiceMock<Server::Configuration::MockFactoryContext> context;
  EXPECT_CALL(context.cluster_manager_.async_client_manager_, factoryForGrpcService(_, _, _))
          .WillOnce(Invoke([](const envoy::api::v2::core::GrpcService &, Stats::Scope &, bool) {
              return std::make_unique<NiceMock<Grpc::MockAsyncClientFactory>>();
          }));
  // when
  auto cb = factory.createFilterFactoryFromProto(*proto_config, "stats", context);
  // then
  EXPECT_NE(cb, nullptr);

  // expect
  StrictMock<Http::MockFilterChainFactoryCallbacks> filter_callback;
  EXPECT_CALL(filter_callback, addStreamDecoderFilter(Pointee(WhenDynamicCastTo<const Filter&>(_))));
  // when
  cb(filter_callback);
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
