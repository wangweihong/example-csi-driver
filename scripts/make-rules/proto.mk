# ==============================================================================
# Makefile helper functions for generate proto files
#


.PHONY: proto.gen
proto.gen: tools.verify.protoc
	@echo "===========> Generating protoc code :${ROOT_DIR}/internal/pkg/genericgrpc/grpcproto/proto"
	@${ROOT_DIR}/scripts/genprotos.sh generate_protos ${ROOT_DIR}/internal/pkg/genericgrpc/grpcproto/proto