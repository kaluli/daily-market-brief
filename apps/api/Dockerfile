FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY apps/api /app
RUN go build -o /server ./cmd/server
RUN go build -o /migrate ./cmd/migrate
RUN go build -o /ingest ./cmd/ingest
RUN go build -o /summarize ./cmd/summarize

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server /app/server
COPY --from=builder /migrate /app/migrate
COPY --from=builder /ingest /app/ingest
COPY --from=builder /summarize /app/summarize
COPY config /app/config
RUN mkdir -p /app/summaries
ENV SUMMARIES_PATH=/app/summaries
ENV PORT=3090
EXPOSE 3090
CMD ["/app/server"]
