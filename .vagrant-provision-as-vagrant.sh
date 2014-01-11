#!/bin/bash

set -e
set -x

ln -svf /vagrant/.vagrant-skel/bashrc ~/.bashrc
ln -svf /vagrant/.vagrant-skel/bash_profile ~/.bash_profile

source ~/.bashrc

go get -x github.com/kr/godep
