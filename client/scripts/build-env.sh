#!/bin/sh

if [ -d dist ] 
then
    file="dist/config.js"
else
    file="config.js"
fi

if [ -z ${VITE_API_ENDPOINT+x} ]
then
    echo 'Error! $VITE_API_ENDPOINT is required.'
    exit 64
fi

if [ -z ${STRIPE_PUBLISHER_KEY+x} ]
then
    echo 'Error! $STRIPE_PUBLISHER_KEY is required.'
    exit 64
fi


configs="
window.configs = window.configs || {};
window.configs.vite_app_endpoint = '$VITE_API_ENDPOINT';
window.configs.stripe_publisher_key = '$STRIPE_PUBLISHER_KEY';
"

if [ -e $file ]
then
    rm $file
fi

echo $configs > $file