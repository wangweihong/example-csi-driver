#!/usr/bin/env bash

mkdir -p "$DST_DIR/kubernetes"




output_file="combined.yaml"

# 删除已存在的输出文件
rm -f "$output_file"

# 合并*.yaml文件到同一个文件
for yaml_file in $(find . -name "*.yaml" ! -name "kustomize.yaml"); do
    echo "---" >> "$output_file"
    cat "$yaml_file" >> "$output_file"
done