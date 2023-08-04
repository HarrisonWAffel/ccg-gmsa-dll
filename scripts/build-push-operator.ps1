cd ../account-provider

docker build . -f Dockerfile.windows -t harrisonwaffel/gmsa:latest

docker push harrisonwaffel/gmsa:latest

cd ../scripts