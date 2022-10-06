#!/usr/bin/bash
oses=(windows darwin linux)
archs=(amd64 arm64)

for os in ${oses[@]}
do
  extension=$([ $os == "windows" ] && echo ".exe" || echo "")
  for arch in ${archs[@]}
  do
    env GOOS=${os} GOARCH=${arch} go build -o dist/tx_${os}_${arch}${extension}
  done
done
