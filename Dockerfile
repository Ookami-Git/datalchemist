FROM alpine:latest

WORKDIR /app

# Copie du binaire correspondant
COPY dist/datalchemist_linux_amd64_v1/datalchemist /usr/local/bin/datalchemist
RUN chmod +x /usr/local/bin/datalchemist

EXPOSE 8080

CMD ["datalchemist"]