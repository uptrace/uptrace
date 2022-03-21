#!/bin/bash

files=(
    bin/uptrace_darwin_arm64
    bin/uptrace_darwin_amd64
    bin/uptrace_linux_arm64
    bin/uptrace_linux_amd64
    bin/uptrace_windows_amd64.exe
    dist/uptrace-*.aarch64.rpm
    dist/uptrace-*.x86_64.rpm
    dist/uptrace_*_amd64.deb
    dist/uptrace_*_arm64.deb
);

for f in "${files[@]}"
do
    if [[ ! -f $f ]]
    then
        echo "$f does not exist."
        echo "::set-output name=passed::false"
        exit 0
    fi
done

echo "::set-output name=passed::true"
