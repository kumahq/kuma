#include "test/integration/http_integration.h"

namespace Envoy {
class KonvoyIntegrationTest : public HttpIntegrationTest,
                              public testing::TestWithParam<Network::Address::IpVersion> {
public:
  KonvoyIntegrationTest()
      : HttpIntegrationTest(Http::CodecClient::Type::HTTP1, GetParam(), realTime()) {}
  /**
   * Initializer for an individual integration test.
   */
  void SetUp() override { initialize(); }

  void initialize() override {
    config_helper_.addFilter(
        "{ name: konvoy, config: { stat_prefix: demo-grpc-server, grpc_service: { envoy_grpc: { "
        "cluster_name: konvoy_demo_side_car } } } }");
    HttpIntegrationTest::initialize();
  }
};

INSTANTIATE_TEST_CASE_P(IpVersions, KonvoyIntegrationTest,
                        testing::ValuesIn(TestEnvironment::getIpVersionsForTest()));

TEST_P(KonvoyIntegrationTest, Test1) {
  Http::TestHeaderMapImpl headers{{":method", "GET"}, {":path", "/"}, {":authority", "host"}};

  IntegrationCodecClientPtr codec_client;
  FakeHttpConnectionPtr fake_upstream_connection;
  FakeStreamPtr request_stream;

  codec_client = makeHttpConnection(lookupPort("http"));
  auto response = codec_client->makeHeaderOnlyRequest(headers);

  /* TODO(yskopets): Fix it
  ASSERT_TRUE(fake_upstreams_[0]->waitForHttpConnection(*dispatcher_, fake_upstream_connection,
                                                        std::chrono::milliseconds(5)));
  ASSERT_TRUE(fake_upstream_connection->waitForNewStream(*dispatcher_, request_stream));
  ASSERT_TRUE(request_stream->waitForEndStream(*dispatcher_));
  */

  response->waitForEndStream();

  codec_client->close();
}
} // namespace Envoy
