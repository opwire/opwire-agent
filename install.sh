#!/usr/bin/env bash

set -e

actions=(download install)

# check for help flag
if [ -n "$1" ] && [[ ! ${actions[*]} =~ "$1" ]]; then
  echo "Usage: curl https://opwire.org/opwire-agent/install.sh | sudo bash" 1>&2;
  exit 1;
fi

# determine the action
task_action=download
if [ -n "$1" ] && [[ ${actions[*]} =~ "$1" ]]; then
  task_action="$1"
fi

# determine current working dir
task_cwd=`pwd`

# determine if the current user has write permission on current working dir
if [ $task_action == 'download' ] && [ ! -w $task_cwd ]; then
  echo "Error: the current user has not write permission on [$task_cwd]"
  exit 5
fi

# create tmp directory and move to it
tmp_dir=`mktemp -d 2>/dev/null || mktemp -d -t 'opwire-install'`; cd $tmp_dir

# make sure unzip tool is available and choose one to work with
unzip_tools_list=('unzip' '7z' 'busybox')
set +e
for tool in ${unzip_tools_list[*]}; do
  trash=`hash $tool 2>>errors`
  if [ "$?" -eq 0 ]; then
    unzip_tool="$tool"
    break
  fi
done
set -e

# exit if no unzip tools available
if [ -z "${unzip_tool}" ]; then
  printf "\nNone of the supported tools for extracting zip archives (${unzip_tools_list[*]}) were found. "
  printf "\nPlease install one of them and try again.\n\n"
  exit 4
fi

#detect the platform
OS="`uname`"
case $OS in
  Linux)
    OS='linux'
    ;;
  FreeBSD)
    OS='freebsd'
    ;;
  NetBSD)
    OS='netbsd'
    ;;
  OpenBSD)
    OS='openbsd'
    ;;
  Darwin)
    OS='darwin'
    ;;
  SunOS)
    OS='solaris'
    echo 'OS not supported'
    exit 2
    ;;
  *)
    echo 'OS not supported'
    exit 2
    ;;
esac

OS_type="`uname -m`"
case $OS_type in
  x86_64|amd64)
    OS_type='amd64'
    ;;
  i?86|x86)
    OS_type='386'
    ;;
  arm*)
    OS_type='arm'
    ;;
  aarch64)
    OS_type='arm64'
    ;;
  *)
    echo 'OS type not supported'
    exit 2
    ;;
esac

# get the latest version
version=`curl -s https://api.github.com/repos/opwire/opwire-agent/releases/latest \
 | sed -n 's/.*"tag_name":.*"\([^"]*\)".*/\1/p'`

if [ -z $version ]; then
  echo "Error: unable to determine the latest version of opwire-agent"
  exit 5
fi

# build download link
zip_file="opwire-agent-$version-$OS-$OS_type.zip"
download_link="https://github.com/opwire/opwire-agent/releases/download/$version/$zip_file"

curl -L -O $download_link
unzip_dir="tmp_unzip_dir_for_opwire"

# select an unzip tool from unzip_tools_list
case $unzip_tool in
  'unzip')
    unzip -a $zip_file -d $unzip_dir
    ;;
  '7z')
    7z x $zip_file -o$unzip_dir
    ;;
  'busybox')
    mkdir -p $unzip_dir
    busybox unzip $zip_file -d $unzip_dir
    ;;
esac

cd $unzip_dir

footer_note="Check https://github.com/opwire/opwire-agent for more details."

if [ $task_action == 'download' ]; then
  cp opwire-agent $task_cwd
  printf "\n${version} has been successfully downloaded."
  printf "\n${footer_note}\n\n"
  exit 0
fi

printf "\nTask[$task_action] has not been implemented"
printf "\n${footer_note}\n\n"
exit 0
