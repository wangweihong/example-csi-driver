#!/usr/bin/env bash

# 可以通过make INSTALL_DIR=xxx的方式设置INSTALL_DIR的值, 其他变量同理。

# 项目源码根目录
SOURCE_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
# 生成文件存放目录
# 如果未指定变量OUT_DIR, 则采用默认值_output
LOCAL_OUTPUT_ROOT="${SOURCE_ROOT}/${OUT_DIR:-_output}"


# 设置安装目录
# 如果未指定变量INSTALL_DIR, 则采用默认值/tmp/installation
readonly INSTALL_DIR=${INSTALL_DIR:-/tmp/installation}
mkdir -p ${INSTALL_DIR}
readonly ENV_FILE=${SOURCE_ROOT}/scripts/install/environment.sh

# eazycloud 配置
readonly EAZYCLOUD_ROOT_DIR=${EAZYCLOUD_ROOT_DIR:-/var/lib/eazycloud}
readonly EAZYCLOUD_DATA_DIR=${EAZYCLOUD_DATA_DIR:-${EAZYCLOUD_ROOT_DIR}/data} # eazycloud 各组件数据目录
readonly EAZYCLOUD_INSTALL_DIR=${EAZYCLOUD_INSTALL_DIR:-${EAZYCLOUD_ROOT_DIR}/bin} # eazycloud 安装文件存放目录
readonly EAZYCLOUD_CONFIG_DIR=${EAZYCLOUD_CONFIG_DIR:-${EAZYCLOUD_ROOT_DIR}/conf} # eazycloud 配置文件存放目录
readonly EAZYCLOUD_LOG_DIR=${EAZYCLOUD_LOG_DIR:-/var/log/eazycloud} # eazycloud 日志文件存放目录
readonly EAZYCLOUD_DEBUG_DIR=${EAZYCLOUD_DEBUG_DIR:-${EAZYCLOUD_ROOT_DIR}/debug} # eazycloud 调试信息文件存放目录
readonly CA_FILE=${CA_FILE:-${EAZYCLOUD_CONFIG_DIR}/cert/ca.pem} # ca


# example-csi-driver配置
## UnixSocket
readonly EXAMPLE_GRPC_RUNTIME_DEBUG_OUTPUT_DIR=${EXAMPLE_GRPC_RUNTIME_DEBUG_OUTPUT_DIR:-${EAZYCLOUD_DEBUG_DIR}/example-csi-driver}
readonly EXAMPLE_CSI_DRIVER_UNIX_SOCKET=${EXAMPLE_CSI_DRIVER_UNIX_SOCKET:-/var/lib/kubelet/plugins/example-csi-driver/csi.sock}

