# go-url-shortener

Production-ready URL kısaltma servisi. Go 1.22, chi router, PostgreSQL ve Redis tabanlı. Clean architecture prensipleri ile yazıldı; handler / service / repository katmanları net biçimde ayrıldı.

## Özellikler

- HTTP API (`POST /api/v1/links`, `GET /:code`, `GET /api/v1/links`, `DELETE /api/v1/links/:code`)
- JWT tabanlı kimlik doğrulama, refresh token desteği
- Hashids ile kısa kod üretimi (collision retry)
- Redis ile sıcak link önbelleği (TTL 1 saat)
- Postgres üzerinde tıklama sayacı (`UPDATE ... RETURNING`)
- Token bucket rate limiter (IP başına)
- Yapılandırılabilir log seviyesi (zerolog)
- `golang-migrate` ile şema yönetimi
- Docker Compose ile tek komutla ayağa kalkar
- GitHub Actions CI: lint + test + build

## Hızlı başlangıç

```bash
cp .env.example .env
docker compose up -d postgres redis
make migrate-up
make run
```

Servis varsayılan olarak `http://localhost:8080` üzerinde çalışır.

## API örnekleri

```bash
# Kayıt ol
curl -X POST localhost:8080/api/v1/auth/register \
  -H "content-type: application/json" \
  -d '{"email":"a@b.com","password":"secret123"}'

# Link oluştur
curl -X POST localhost:8080/api/v1/links \
  -H "authorization: Bearer <token>" \
  -H "content-type: application/json" \
  -d '{"url":"https://example.com/very/long/path"}'

# Yönlendirme
curl -L localhost:8080/abc123
```

## Geliştirme

```bash
make test           # unit testler
make test-int       # integration testler (postgres + redis gerekir)
make lint           # golangci-lint
make build          # ./bin/server
```

## Mimari

```
cmd/server          → entrypoint, dependency wiring
internal/handler    → HTTP layer, request/response DTO'ları
internal/service    → iş kuralları, transaction yönetimi
internal/repository → Postgres / Redis erişim katmanı
internal/domain     → domain modelleri ve hatalar
internal/pkg        → reusable yardımcılar (jwt, hashids)
```

## Lisans

MIT
