FROM scratch
LABEL maintainer "Seth Vargo <seth@sethvargo.com> (@sethvargo)"

ADD "https://curl.se/ca/cacert.pem" "/etc/ssl/certs/ca-certificates.crt"
ADD "http-echo" "/"
ENTRYPOINT ["/http-echo"]
