FROM alpine:latest

WORKDIR /app

# Set the target architecture
ARG TARGETARCH

# Copie du binaire correspondant
COPY linux/${TARGETARCH}/datalchemist /usr/local/bin/datalchemist
RUN chmod +x /usr/local/bin/datalchemist

EXPOSE 8080

CMD ["datalchemist"]