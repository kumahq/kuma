#include "extensions/filters/http/konvoy/konvoy.h"

#include "common/protobuf/utility.h"

#include "test/mocks/grpc/mocks.h"
#include "test/mocks/http/mocks.h"
#include "test/mocks/stats/mocks.h"
#include "test/test_common/simulated_time_system.h"
#include "test/test_common/utility.h"

#include "gmock/gmock.h"
#include "gtest/gtest.h"

using KonvoyProtoConfig = envoy::config::filter::http::konvoy::v2alpha::Konvoy;
using envoy::service::konvoy::v2alpha::ProxyHttpRequestClientMessage;
using envoy::service::konvoy::v2alpha::ProxyHttpRequestServerMessage;
using testing::NiceMock;
using testing::ReturnRef;
using testing::StrictMock;
using testing::WhenDynamicCastTo;

namespace Envoy {
namespace Extensions {
namespace HttpFilters {
namespace Konvoy {

class KonvoyHttpFilterTest : public testing::Test {
public:
    ConfigSharedPtr setupConfig(std::function<void(KonvoyProtoConfig&)> configure) {
      KonvoyProtoConfig config;
      config.set_stat_prefix("demo-grpc-server.");
      if (configure) {
        configure(config);
      }
      return std::make_shared<Config>(config, store_, time_system_);
    }

    KonvoyHttpFilterTest() : config_(setupConfig(nullptr)), async_client_(new StrictMock<Grpc::MockAsyncClient>()) {
      filter_ = std::make_unique<Filter>(config_, Grpc::AsyncClientPtr{async_client_});
      filter_->setDecoderFilterCallbacks(callbacks_);
    }

    Stats::IsolatedStoreImpl store_;
    Event::SimulatedTimeSystem time_system_;
    ConfigSharedPtr config_;
    StrictMock<Grpc::MockAsyncClient>* async_client_;
    StrictMock<Grpc::MockAsyncStream> async_stream_;

    StrictMock<Http::MockStreamDecoderFilterCallbacks> callbacks_;
    Buffer::OwnedImpl data_;

    std::unique_ptr<Filter> filter_;
    Http::TestHeaderMapImpl headers_;
};

TEST_F(KonvoyHttpFilterTest, KonvoyServiceIsDown) {
  // expect : an attempt to open a new stream to Http Konvoy Service to fail
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              // simulate unavailable cluster
              filter_->onRemoteClose(Grpc::Status::GrpcStatus::Unavailable, "Cluster not available");
              return nullptr;
          }));
  // and expect : response with HTTP 500
  Http::TestHeaderMapImpl response_headers{{":status", "500"}};
  EXPECT_CALL(callbacks_, encodeHeaders_(HeaderMapEqualRef(&response_headers), true));
  // when : HTTP request headers received
  EXPECT_EQ(Http::FilterHeadersStatus::StopIteration, filter_->decodeHeaders(headers_, true));
  // then : metrics get updated
  EXPECT_EQ(0U, config_->stats().rq_active_.value());
  EXPECT_EQ(1U, config_->stats().rq_total_.value());
  EXPECT_EQ(1U, config_->stats().rq_error_.value());
  EXPECT_EQ(0U, config_->stats().rq_cancel_.value());
}

} // namespace Konvoy
} // namespace HttpFilters
} // namespace Extensions
} // namespace Envoy
