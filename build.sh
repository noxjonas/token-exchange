#!/usr/bin/bash
oses=(windows darwin linux)
archs=(amd64 arm64)

for os in ${oses[@]}
do
  for arch in ${archs[@]}
  do
    env GOOS=${os} GOARCH=${arch} go build -o dist/tx_${os}_${arch}
  done
done


#
#      # Runs a set of commands using the runners shell
#      - name: Run a multi-line script
#        run: |
#          oses=(windows darwin linux)
#          archs=(amd64 arm64)
#
#          for os in ${oses[@]}
#          do
#            for arch in ${archs[@]}
#            do
#              env GOOS=${os} GOARCH=${arch} go build -o dist/tx_${os}_${arch}
#            done
#          done