#!/bin/sh

if [ -d dist ] 
then
    file="dist/config.js"
else
    file="config.js"
fi

# if [ -z ${VITE_API_ENDPOINT+x} ]
# then
#     echo 'Error! $VITE_API_ENDPOINT is required.'
#     exit 64
# fi


configs="
window.configs = window.configs || {};
window.configs.vite_app_endpoint = 'http://localhost:3000/v1';
"

if [ -e $file ]
then
    rm $file
fi

echo $configs > $file