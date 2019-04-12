#include "extensions/filters/network/konvoy/konvoy.h"

#include <string>

#include "common/buffer/buffer_impl.h"

#include "extensions/filters/network/konvoy/proto_utils.h"

namespace Envoy {
namespace Extensions {
namespace NetworkFilters {
namespace Konvoy {

InstanceStats Config::generateStats(const std::string& name, Stats::Scope& scope) {
  const std::string final_prefix = fmt::format("konvoy.network.{}.", name);
  return {ALL_TCP_KONVOY_STATS(
          POOL_GAUGE_PREFIX(scope, final_prefix),
          POOL_COUNTER_PREFIX(scope, final_prefix),
          POOL_HISTOGRAM_PREFIX(scope, final_prefix))};
}

Config::Config(
    const envoy::config::filter::network::konvoy::v2alpha::Konvoy& config,
    Stats::Scope& scope, TimeSource& time_source)
    : proto_config_(config), stats_(generateStats(config.stat_prefix(), scope)), time_source_(time_source),
      scope_(scope) {}

Filter::Filter(ConfigSharedPtr config, Grpc::AsyncClientPtr&& async_client)
    : config_(config),
    async_client_(std::move(async_client)),
    service_method_(*Protobuf::DescriptorPool::generated_pool()->FindMethodByName("envoy.service.konvoy.v2alpha.NetworkKonvoy.ProxyConnection")) {}

Filter::~Filter() {}

// Network::ReadFilter

void Filter::initializeReadFilterCallbacks(Network::ReadFilterCallbacks& callbacks) {
  read_callbacks_ = &callbacks;
  read_callbacks_->connection().addConnectionCallbacks(*this);
}

Network::FilterStatus Filter::onNewConnection() {
  return Network::FilterStatus::Continue;
}

Network::FilterStatus Filter::onData(Buffer::Instance& data, bool end_stream) {
  ENVOY_LOG_MISC(trace, "konvoy-network-filter: forwarding request data to Network Konvoy Service (side car):\n{} bytes, end_stream={}",
                 data.length(), end_stream);

  if (state_ == State::NotStarted) {
    state_ = State::Streaming;

    config_->stats().cx_total_.inc();

    start_stream_ = config_->timeSource().monotonicTime();

    // need to increment in advance to support a scenario where `onRemoteClose` is called back from `start`
    config_->stats().cx_active_.inc();

    stream_ = async_client_->start(service_method_, *this);

    if (stream_ == nullptr) {
      ENVOY_LOG_MISC(debug, "konvoy-network-filter: failed to start a new stream to the Network Konvoy Service (side car)");
      // error handling must already have happened inside `onRemoteClose`
      ASSERT(state_ == State::Complete);
      return Network::FilterStatus::StopIteration;
    }

    // remember the buffer for later use in combination with `read_callbacks_->continueReading()`
    buffer_ = &data;

    if (config_->getProtoConfig().per_service_config().has_network_konvoy()) {
      auto &network_konvoy_config = config_->getProtoConfig().per_service_config().network_konvoy();

      auto config_message = KonvoyProtoUtils::serviceConfigurationMessage(network_konvoy_config);

      stream_->sendMessage(config_message, false);
    }
  }
  if (state_ == State::Complete) {
    return Network::FilterStatus::Continue;
  }
  
  auto message = KonvoyProtoUtils::requestDataChunkMessage(data);

  stream_->sendMessage(message, end_stream);

  endStreamIfNecessary(end_stream);

  // drain the buffer before passing control to the next filter
  data.drain(data.length());

  // Konvoy Network Filter is expected to be used in one of the following ways:
  // 1) as a terminal filter in the chain
  // 2) as an intermediate filter followed by `envoy.tcp_proxy`
  // Thus, we could return here either `StopIteration`
  // or `Continue` with 0-length buffer.
  // We do return `Continue` to let `envoy.tcp_proxy` establish
  // connection to upstream as soon as possible.
  return Network::FilterStatus::Continue;
}

// Network::ConnectionCallbacks

void Filter::onEvent(Network::ConnectionEvent event) {
  if (event == Network::ConnectionEvent::RemoteClose ||
      event == Network::ConnectionEvent::LocalClose) {
    if (state_ == State::Streaming) {
      state_ = State::Complete;
      stream_->resetStream();
      config_->stats().cx_active_.dec();
      chargeStreamStats(Grpc::Status::GrpcStatus::Canceled);
    }
  }
}

// Grpc::AsyncStreamCallbacks

void Filter::onReceiveMessage(std::unique_ptr<envoy::service::konvoy::v2alpha::ProxyConnectionServerMessage>&& message) {
  ENVOY_LOG_MISC(trace, "konvoy-network-filter: received message from Network Konvoy Service (side car):\n{}", message->message_case());

  switch (message->message_case()) {
    // Network Konvoy Service (side car) modified the original request data and wants us to pass control to the next filter in the chain
    case envoy::service::konvoy::v2alpha::ProxyConnectionServerMessage::MessageCase::kRequestDataChunk: {

      if (0 < buffer_->length()) {
        ENVOY_LOG_MISC(error, "konvoy-network-filter: buffer is expected to be empty when response from gRPC service is received, but got {} bytes instead:\n{}", buffer_->length(), buffer_->toString());
        break;
      }

      buffer_->add(message->request_data_chunk().bytes());
      read_callbacks_->continueReading();

      if (0 < buffer_->length()) {
        ENVOY_LOG_MISC(error, "konvoy-network-filter: buffer is expected to be empty once call to `read_callbacks_->continueReading()` completes, but got {} bytes instead:\n{}", buffer_->length(), buffer_->toString());
      }
      break;
    }
    // Network Konvoy Service (side car) returned response data and wants us to forward them to downstream verbatim
    case envoy::service::konvoy::v2alpha::ProxyConnectionServerMessage::MessageCase::kResponseDataChunk: {
      Buffer::OwnedImpl data{message->response_data_chunk().bytes()};
      read_callbacks_->connection().write(data, false);
      ASSERT(0 == data.length());
      break;
    }
    default:
      break;
  }
}

void Filter::onRemoteClose(Grpc::Status::GrpcStatus status, const std::string& message) {
  ENVOY_LOG_MISC(trace, "konvoy-network-filter: received close signal from Network Konvoy Service (side car):\nstatus = {}, message = {}", status, message);

  state_ = State::Complete;
  config_->stats().cx_active_.dec();

  if (status == Grpc::Status::GrpcStatus::Ok) {
    read_callbacks_->continueReading();
  } else {
    config_->stats().cx_error_.inc();
    read_callbacks_->connection().close(Network::ConnectionCloseType::NoFlush);
  }
}

void Filter::endStreamIfNecessary(bool end_stream) {
  if (end_stream) {
    endStream();
  }
}

void Filter::endStream() {
  stream_->closeStream();
}

void Filter::chargeStreamStats(Grpc::Status::GrpcStatus) {
  if (!stream_) {
    // if a stream has never been opened, stats make no sense
    return;
  }

  auto now = config_->timeSource().monotonicTime();

  std::chrono::milliseconds totalLatency = std::chrono::duration_cast<std::chrono::milliseconds>(now - start_stream_);

  config_->stats().cx_stream_latency_ms_.recordValue(totalLatency.count());
  config_->stats().cx_total_stream_latency_ms_.add(totalLatency.count());
}

} // namespace Konvoy
} // namespace NetworkFilters
} // namespace Extensions
} // namespace Envoy
