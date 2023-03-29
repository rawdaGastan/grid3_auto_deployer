#!/bin/sh
if [ -z ${VITE_API_ENDPOINT+x} ]
then
    echo 'Error! $VITE_API_ENDPOINT is required.'
    exit 64
else
    
    yarn run build
    sh ./scripts/build-env.sh
    cd dist
    npx lite-server
fi