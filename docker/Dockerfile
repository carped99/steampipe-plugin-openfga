# syntax=docker/dockerfile:1.19.0

# ------------------------------------------------------------
# 1단계: steampipe-postgres-fdw + plugin 을 빌드하는 builder
# ------------------------------------------------------------
ARG GO_VERSION=1.25.4
ARG POSTGIS_TAG=17-3.5
FROM golang:${GO_VERSION} AS go-builder

# -------------------------------------------------------
# 3. Clone steampipe-postgres-fdw
# -------------------------------------------------------
WORKDIR /src
RUN git clone --depth 1 https://github.com/turbot/steampipe-postgres-fdw.git
WORKDIR /src/steampipe-postgres-fdw

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

FROM postgis/postgis:${POSTGIS_TAG} AS fdw-builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

# -------------------------------------------------------
# 1. System packages + PostgreSQL dev files
# -------------------------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates \
    postgresql-server-dev-${PG_MAJOR} \
    rsync git wget curl \
    && rm -rf /var/lib/apt/lists/*

# 2-2. go-src 스테이지에서 Go 바이너리 복사 (wget X)
COPY --from=go-builder /usr/local/go /usr/local/go
ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod

WORKDIR /src/steampipe-postgres-fdw
COPY --from=go-builder /src/steampipe-postgres-fdw ./

# -------------------------------------------------------
# 4. Build FDW
# -------------------------------------------------------
ARG PLUGIN_NAME=openfga
ARG PLUGIN_VERSION="v0.0.1"
ARG PLUGIN_GITHUB_URL=github.com/carped99/steampipe-plugin-openfga

ENV GOFLAGS="-trimpath -ldflags=-s -ldflags=-w"

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make standalone \
      plugin="${PLUGIN_NAME}" \
      plugin_version="${PLUGIN_VERSION}" \
      plugin_github_url="${PLUGIN_GITHUB_URL}"

FROM postgis/postgis:${POSTGIS_TAG} AS runtime
ARG PG_MAJOR

# Copy FDW extension files
COPY --from=fdw-builder /src/steampipe-postgres-fdw/build-Linux/steampipe_postgres_openfga.so /usr/lib/postgresql/${PG_MAJOR}/lib/openfga_fdw.so
#COPY --from=fdw-builder /src/steampipe-postgres-fdw/build-Linux/steampipe_postgres_openfga.control /usr/share/postgresql/${PG_MAJOR}/extension/
#COPY --from=fdw-builder /src/steampipe-postgres-fdw/build-Linux/steampipe_postgres_openfga--*.sql /usr/share/postgresql/${PG_MAJOR}/extension/

# Copy extension metadata from local extensions directory
COPY extensions/openfga_fdw.control /usr/share/postgresql/${PG_MAJOR}/extension/
COPY extensions/openfga_fdw--*.sql /usr/share/postgresql/${PG_MAJOR}/extension/

# Copy initialization script to auto-install extension
#COPY docker/initdb-openfga.sh /docker-entrypoint-initdb.d/10_openfga.sh
#RUN chmod +x /docker-entrypoint-initdb.d/10_openfga.sh

