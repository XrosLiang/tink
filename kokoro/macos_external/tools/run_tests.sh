#!/bin/bash

set -euo pipefail
cd ${KOKORO_ARTIFACTS_DIR}/git/tink
cd tools

# TODO(b/155225382): Avoid modifying the sytem Python installation.
pip3 install --user protobuf

use_bazel.sh $(cat .bazelversion)
time bazel build -- ...
time bazel test --test_output=errors -- ...
