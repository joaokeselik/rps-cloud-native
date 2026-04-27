FROM python:3.12-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends postgresql-client ca-certificates && \
    pip install --no-cache-dir awscli && \
    rm -rf /var/lib/apt/lists/*

COPY backup/postgres-to-s3.sh /usr/local/bin/postgres-to-s3.sh
RUN chmod +x /usr/local/bin/postgres-to-s3.sh

ENTRYPOINT ["/usr/local/bin/postgres-to-s3.sh"]
