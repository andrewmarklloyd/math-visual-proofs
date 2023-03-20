#!/bin/bash


scriptpath=${1}
appname=${2}

docker run --rm -it --user="$(id -u):$(id -g)" -v "$(pwd)":/manim manimcommunity/manim:stable manim ${scriptpath} ${appname} -qm

aws s3 --endpoint=${SPACES_URL} cp ./media/videos/${appname}/720p30/${appname}.mp4 s3://math-visual-proofs/renderings/
