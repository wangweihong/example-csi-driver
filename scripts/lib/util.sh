#!/usr/bin/env bash

# Test whether openssl is installed.
# Sets:
#  OPENSSL_BIN: The path to the openssl binary to use
function lib::util::test_openssl_installed {
    if ! openssl version >& /dev/null; then
      echo "Failed to run openssl. Please ensure openssl is installed"
      exit 1
    fi

    OPENSSL_BIN=$(command -v openssl)
}

# Test whether kustomize is installed.
# Sets:
#  KUSTOMIZE_BIN: The path to the kustomize binary to use
function lib::util::test_kustomize_installed {
    if ! kustomize version >& /dev/null; then
      echo "Failed to run kustomize. Please ensure kustomize is installed"
      exit 1
    fi

    KUSTOMIZE_BIN=$(command -v kustomize)
}