# use lightweight alpine for now
FROM alpine:latest

# install needed applications
# RUN apk add --no-cache git curl

# download k3s binary and script
# RUN curl -Lo /usr/local/bin/k3s https://github.com/k3s-io/k3s/releases/download/v1.31.6+k3s1/k3s 
# RUN chmod a+x /usr/local/bin/k3s

# custom env value
ENV _NO_K3S_DOWNLOAD=true
