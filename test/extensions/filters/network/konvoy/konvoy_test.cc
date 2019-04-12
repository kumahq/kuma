#include "extensions/filters/network/konvoy/konvoy.h"

#include "test/mocks/grpc/mocks.h"
#include "test/mocks/network/mocks.h"
#include "test/test_common/simulated_time_system.h"

#include "gmock/gmock.h"
#include "gtest/gtest.h"

using envoy::service::konvoy::v2alpha::ProxyConnectionClientMessage;
using envoy::service::konvoy::v2alpha::ProxyConnectionServerMessage;
using testing::NiceMock;
using testing::ReturnRef;
using testing::StrictMock;
using testing::WhenDynamicCastTo;

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

class KonvoyFilterTest : public testing::Test {
public:
    ConfigSharedPtr setupConfig() {
      envoy::config::filter::network::konvoy::v2alpha::Konvoy config;
      config.set_stat_prefix("demo-grpc-server.");
      return std::make_shared<Config>(config, store_, time_system_);
    }

    KonvoyFilterTest() : config_(setupConfig()), async_client_(new StrictMock<Grpc::MockAsyncClient>()) {
      EXPECT_CALL(callbacks_, connection()).WillRepeatedly(ReturnRef(callbacks_.connection_));

      filter_ = std::make_unique<Filter>(config_, Grpc::AsyncClientPtr{async_client_});
      filter_->initializeReadFilterCallbacks(callbacks_);
    }

    Stats::IsolatedStoreImpl store_;
    Event::SimulatedTimeSystem time_system_;
    ConfigSharedPtr config_;
    StrictMock<Grpc::MockAsyncClient>* async_client_;
    StrictMock<Grpc::MockAsyncStream> async_stream_;

    StrictMock<Network::MockReadFilterCallbacks> callbacks_;
    Buffer::OwnedImpl data_;

    std::unique_ptr<Filter> filter_;
};

TEST_F(KonvoyFilterTest, ConnectionWithoutPayload) {
  // when
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(0U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // when
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::RemoteClose));

  // then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(0U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, KonvoyServiceIsDown) {
  // when
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given
  data_.add("Hello");
  // expect
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              // simulate unavailable cluster
              filter_->onRemoteClose(Grpc::Status::GrpcStatus::Unavailable, "Cluster not available");
              return nullptr;
          }));
  // and expect
  EXPECT_CALL(callbacks_.connection_, close(Network::ConnectionCloseType::NoFlush));

  // when
  EXPECT_EQ(Network::FilterStatus::StopIteration, filter_->onData(data_, false));

  // then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(1U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
  EXPECT_EQ("Hello", data_.toString());

  // given
  data_.add(" world!");

  // when
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, false));

  // then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(1U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
  EXPECT_EQ("Hello world!", data_.toString());
}

TEST_F(KonvoyFilterTest, MutateSimpleRequest) {
  // expect
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given
  data_.add("Hello");
  // expect
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), true));
  // and expect
  EXPECT_CALL(async_stream_, closeStream());
  // when
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, true));
  // then
  EXPECT_EQ(0U, data_.length());
  // and then
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // given
  auto server_message = std::make_unique<ProxyConnectionServerMessage>();
  server_message->mutable_request_data_chunk()->set_bytes("Hello world!");
  // expect
  EXPECT_CALL(callbacks_, continueReading())
          .WillOnce(Invoke([this]() {
              EXPECT_EQ("Hello world!", this->data_.toString());
              this->data_.drain(this->data_.length());
          }));
  // when
  EXPECT_NO_THROW(filter_->onReceiveMessage(std::move(server_message)));
  // then
  EXPECT_EQ(0U, data_.length());
  // and then
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // expect
  EXPECT_CALL(callbacks_, continueReading())
          .WillOnce(Invoke([this]() {
              EXPECT_EQ(0U, this->data_.length());
          }));
  // when
  EXPECT_NO_THROW(filter_->onRemoteClose(Grpc::Status::GrpcStatus::Ok, "OK"));
  // then
  EXPECT_EQ(0U, data_.length());
  // and then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, RemoteClosesConnection) {
  // expect
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given
  data_.add("Hello");
  // expect
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), false));
  // when
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, false));
  // then
  EXPECT_EQ(0U, data_.length());
  // and then
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // expect
  EXPECT_CALL(async_stream_, resetStream());
  // when
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::RemoteClose));
  // then
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(1U, config_->stats().cx_cancel_.value());

  // when : unexpected subsequent invocation
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::LocalClose));
  // then : must be no-op
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(1U, config_->stats().cx_cancel_.value());
}

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
