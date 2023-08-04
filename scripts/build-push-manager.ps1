cd ../dll-installer

docker build . -f Dockerfile.windows -t harrisonwaffel/gmsa-plugin-manager:latest

docker push harrisonwaffel/gmsa-plugin-manager:latest

cd ../scripts