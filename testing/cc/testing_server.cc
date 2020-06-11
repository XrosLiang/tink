// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
///////////////////////////////////////////////////////////////////////////////

#include <grpcpp/grpcpp.h>

#include "absl/flags/flag.h"
#include "absl/flags/parse.h"
#include "tink/config/tink_config.h"
#include "proto/testing/testing_api.grpc.pb.h"
#include "aead_impl.h"
#include "keyset_impl.h"

ABSL_FLAG(int, port, 23456, "the port");

void RunServer() {
  auto status = crypto::tink::TinkConfig::Register();
  if (!status.ok()) {
    std::cout << "TinkConfig::Register() failed: " << status.error_message()
              << std::endl;
    return;
  }
  const int port = absl::GetFlag(FLAGS_port);
  std::string server_address = absl::StrCat("[::]:", port);

  tink_testing_api::KeysetImpl keyset;
  tink_testing_api::AeadImpl aead;

  grpc::ServerBuilder builder;
  builder.AddListeningPort(
      server_address, ::grpc::experimental::LocalServerCredentials(LOCAL_TCP));

  builder.RegisterService(&keyset);
  builder.RegisterService(&aead);

  std::unique_ptr<grpc::Server> server(builder.BuildAndStart());
  std::cout << "Server listening on " << server_address << std::endl;
  server->Wait();
}

int main(int argc, char** argv) {
  absl::ParseCommandLine(argc, argv);
  RunServer();
  return 0;
}