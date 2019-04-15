#include "api/envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.h"
#include "api/envoy/config/filter/network/konvoy/v2alpha/konvoy.pb.validate.h"

#include "envoy/stats/scope.h"

#include "extensions/filters/network/konvoy/config.h"
#include "extensions/filters/network/konvoy/konvoy.h"

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
namespace NetworkFilters {
namespace Konvoy {

TEST(KonvoyFilterConfigFactoryTest, InvalidProto) {
  // given
  envoy::config::filter::network::konvoy::v2alpha::Konvoy proto_config;

  // where
  NiceMock<Server::Configuration::MockFactoryContext> context;
  KonvoyFilterConfigFactory factory;

  // expect
  EXPECT_THROW(factory.createFilterFactoryFromProto(
          proto_config, context), ProtoValidationException);
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
  auto cb = factory.createFilterFactoryFromProto(*proto_config, context);
  // then
  EXPECT_NE(cb, nullptr);

  // expect
  StrictMock<Network::MockConnection> filter_manager;
  EXPECT_CALL(filter_manager, addReadFilter(Pointee(WhenDynamicCastTo<const Filter&>(_))));
  // when
  cb(filter_manager);
}

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
