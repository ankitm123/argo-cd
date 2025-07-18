#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -x
set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"/..
  pwd
)
PATH="${PROJECT_ROOT}/dist:${PATH}"
GOPATH=$(go env GOPATH)
GOPATH_PROJECT_ROOT="${GOPATH}/src/github.com/argoproj/argo-cd"

TARGET_SCRIPT=kube_codegen.sh

# codegen utilities are installed outside kube_codegen.sh so remove the `go install` step in the script.
sed -e '/go install/d' "${PROJECT_ROOT}/vendor/k8s.io/code-generator/kube_codegen.sh" > ${TARGET_SCRIPT}

# generate-groups.sh assumes codegen utilities are installed to GOBIN, but we just ensure the CLIs
# are in the path and invoke them without assumption of their location
# shellcheck disable=SC2016
sed -i.bak -e 's#${gobin}/##g' ${TARGET_SCRIPT}

[ -e ./v3 ] || ln -s . v3
[ -e "${GOPATH_PROJECT_ROOT}" ] || (mkdir -p "$(dirname "${GOPATH_PROJECT_ROOT}")" && ln -s "${PROJECT_ROOT}" "${GOPATH_PROJECT_ROOT}")

# shellcheck source=pkg/apis/application/v1alpha1/kube_codegen.sh
. ${TARGET_SCRIPT}

kube::codegen::gen_helpers pkg/apis/application/v1alpha1
kube::codegen::gen_client pkg/apis \
  --output-dir pkg/client \
  --output-pkg github.com/argoproj/argo-cd/v3/pkg/client \
  --boilerplate "${PROJECT_ROOT}/hack/custom-boilerplate.go.txt" \
  --with-watch

rm ${TARGET_SCRIPT}
rm ${TARGET_SCRIPT}.bak

[ -L "${GOPATH_PROJECT_ROOT}" ] && rm -rf "${GOPATH_PROJECT_ROOT}"
[ -L ./v3 ] && rm -rf v3
