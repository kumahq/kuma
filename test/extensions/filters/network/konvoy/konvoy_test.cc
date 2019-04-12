#include "extensions/filters/network/konvoy/konvoy.h"

#include "common/protobuf/utility.h"

#include "test/mocks/grpc/mocks.h"
#include "test/mocks/network/mocks.h"
#include "test/test_common/simulated_time_system.h"

#include "gmock/gmock.h"
#include "gtest/gtest.h"

using KonvoyProtoConfig = envoy::config::filter::network::konvoy::v2alpha::Konvoy;
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
    ConfigSharedPtr setupConfig(std::function<void(KonvoyProtoConfig&)> configure) {
      KonvoyProtoConfig config;
      config.set_stat_prefix("demo-grpc-server.");
      if (configure) {
        configure(config);
      }
      return std::make_shared<Config>(config, store_, time_system_);
    }

    KonvoyFilterTest() : config_(setupConfig(nullptr)), async_client_(new StrictMock<Grpc::MockAsyncClient>()) {
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
  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // then : metrics don't change yet
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(0U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // when : a connection is closed before any data has been received
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::RemoteClose));

  // then : metrics still don't change
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(0U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, KonvoyServiceIsDown) {
  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given : request data frame
  data_.add("Hello");
  // expect : an attempt to open a new stream to Network Konvoy Service to fail
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              // simulate unavailable cluster
              filter_->onRemoteClose(Grpc::Status::GrpcStatus::Unavailable, "Cluster not available");
              return nullptr;
          }));
  // and expect : connection to be closed
  EXPECT_CALL(callbacks_.connection_, close(Network::ConnectionCloseType::NoFlush));

  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::StopIteration, filter_->onData(data_, false));

  // then : metrics get updated
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(1U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
  // then : buffer is not consumed
  EXPECT_EQ("Hello", data_.toString());

  // given : another request data frame
  data_.add(" world!");

  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, false));

  // then : metrics don't change
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(1U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
  // then : buffer is not consumed
  EXPECT_EQ("Hello world!", data_.toString());
}

TEST_F(KonvoyFilterTest, MutateSimpleRequest) {
  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given : request data frame
  data_.add("Hello");
  // expect : an attempt to open a new stream to Network Konvoy Service to succeed
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect : request data frame to be forwarded to Network Konvoy Service
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), true));
  // and expect : stream to Network Konvoy Service to be half closed
  EXPECT_CALL(async_stream_, closeStream());
  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, true));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics get updated
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // given : Network Konvoy Service mutates request data
  auto server_message = std::make_unique<ProxyConnectionServerMessage>();
  server_message->mutable_request_data_chunk()->set_bytes("Hello world!");
  // expect : filter chain to proceed with request data modified by Network Konvoy Service
  EXPECT_CALL(callbacks_, continueReading())
          .WillOnce(Invoke([this]() {
              EXPECT_EQ("Hello world!", this->data_.toString());
              this->data_.drain(this->data_.length());
          }));
  // when : a message with mutated request data is received from Network Konvoy Service
  EXPECT_NO_THROW(filter_->onReceiveMessage(std::move(server_message)));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics don't change
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // expect : filter chain to proceed with empty buffer
  EXPECT_CALL(callbacks_, continueReading())
          .WillOnce(Invoke([this]() {
              EXPECT_EQ(0U, this->data_.length());
          }));
  // when : a stream is closed by Network Konvoy Service
  EXPECT_NO_THROW(filter_->onRemoteClose(Grpc::Status::GrpcStatus::Ok, "OK"));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics get updated
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, DirectResponse) {
  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given : request data frame
  data_.add("Hello");
  // expect : an attempt to open a new stream to Network Konvoy Service to succeed
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect : request data frame to be forwarded to Network Konvoy Service
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), true));
  // and expect : stream to Network Konvoy Service to be half closed
  EXPECT_CALL(async_stream_, closeStream());
  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, true));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics get updated
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // given : Network Konvoy Service responds directly to downstream
  auto server_message = std::make_unique<ProxyConnectionServerMessage>();
  server_message->mutable_response_data_chunk()->set_bytes("Greetings!");
  // and expect : response forwarded to downstream
  EXPECT_CALL(callbacks_.connection_, write(_, false));
  // when : a message with a direct response to downstream is received from Network Konvoy Service
  EXPECT_NO_THROW(filter_->onReceiveMessage(std::move(server_message)));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics don't change
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // expect : a stream to Network Konvoy Service to be terminated
  EXPECT_CALL(async_stream_, resetStream());
  // when : downstream closes connection
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::RemoteClose));
  // and then : metrics get updated
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(1U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, ServiceConfiguration) {
  // given : configuration for Network Konvoy Servivce is defined
  auto service_config = MessageUtil::keyValueStruct("fixed_delay", "2s");
  config_ = setupConfig([&service_config](KonvoyProtoConfig& config) {
    config.mutable_per_service_config()->mutable_network_konvoy()->PackFrom(service_config);
  });
  // and given
  async_client_ = new StrictMock<Grpc::MockAsyncClient>();
  filter_ = std::make_unique<Filter>(config_, Grpc::AsyncClientPtr{async_client_});
  filter_->initializeReadFilterCallbacks(callbacks_);

  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given : request data frame
  data_.add("Hello");
  // expect : an attempt to open a new stream to Network Konvoy Service to succeed
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect : configuraration to be forwarded to Network Konvoy Service
  ProxyConnectionClientMessage configuration_message{};
  configuration_message.mutable_configuration()->mutable_config()->PackFrom(service_config);
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(configuration_message)), false));
  // and expect : request data frame to be forwarded to Network Konvoy Service
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), true));
  // and expect : stream to Network Konvoy Service to be half closed
  EXPECT_CALL(async_stream_, closeStream());
  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, true));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics get updated
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());
}

TEST_F(KonvoyFilterTest, RemoteClosesConnection) {
  // when : a new connection is established
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onNewConnection());

  // given : request data frame
  data_.add("Hello");
  // expect : an attempt to open a new stream to Network Konvoy Service to succeed
  EXPECT_CALL(*async_client_, start(_, _))
          .WillOnce(Invoke([this](const Protobuf::MethodDescriptor&, Grpc::AsyncStreamCallbacks&) {
              return &this->async_stream_;
          }));
  // and expect : request data frame to be forwarded to Network Konvoy Service
  ProxyConnectionClientMessage client_message{};
  client_message.mutable_request_data_chunk()->set_bytes("Hello");
  EXPECT_CALL(async_stream_, sendMessage(WhenDynamicCastTo<const ProxyConnectionClientMessage&>(ProtoEq(client_message)), false));
  // when : data frame is received
  EXPECT_EQ(Network::FilterStatus::Continue, filter_->onData(data_, false));
  // then : buffer is consumed
  EXPECT_EQ(0U, data_.length());
  // and then : metrics get updated
  EXPECT_EQ(1U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(0U, config_->stats().cx_cancel_.value());

  // expect : a stream to Network Konvoy Service to be terminated
  EXPECT_CALL(async_stream_, resetStream());
  // when : downstream closes connection
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::RemoteClose));
  // then : metrics get updated
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(1U, config_->stats().cx_cancel_.value());

  // when : unexpected subsequent invocation
  EXPECT_NO_THROW(filter_->onEvent(Network::ConnectionEvent::LocalClose));
  // then : metrics don't change
  EXPECT_EQ(0U, config_->stats().cx_active_.value());
  EXPECT_EQ(1U, config_->stats().cx_total_.value());
  EXPECT_EQ(0U, config_->stats().cx_error_.value());
  EXPECT_EQ(1U, config_->stats().cx_cancel_.value());
}

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
