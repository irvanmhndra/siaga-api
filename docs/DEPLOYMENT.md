# Undangin Deployment Guide

Diadaptasi dari pola nexpos (VPS + Docker + GHCR + GitHub Actions + nginx + certbot),
dengan beberapa penyesuaian khusus Undangin (lihat **Beda dari nexpos** di bawah).

## Arsitektur

**Single-origin** — beda dari nexpos yang dua origin. Karena Undangin butuh
**Open Graph injection** buat preview WhatsApp di `/i/:slug`, halaman publik HARUS
diserve oleh Go API (bukan nginx statis). Jadi satu container Go melayani SPA + API + OG.

- **undangin.irvanmahendra.com** → nginx → Go API `:8088`
  - Go serve SPA (`WEB_DIST`) + inject OG di `/i/:slug` + endpoint `/api/v1/*`
- Cukup **satu subdomain**. `api.undangin...` tidak diperlukan (hemat 1 record + 1 SAN cert).
- **cdn.undangin.irvanmahendra.com** → R2 (custom domain, sudah di-setup, dikelola Cloudflare)

### Beda dari nexpos (dan alasannya)
| Hal | nexpos | undangin | kenapa |
|---|---|---|---|
| Serve web | nginx statis, origin terpisah | Go serve SPA (single-origin) | butuh OG injection utk WhatsApp preview |
| Subdomain | web + api | cukup `undangin.` | single-origin |
| Auth | JWT (`JWT_SECRET`) | opaque token (tabel `user_sessions`) | tak perlu `JWT_SECRET` |
| Migrasi | step migrate terpisah di CI | auto-migrate saat boot (`app.go`) | tak perlu step CI terpisah |
| Postgres port | `5432:5432` (ter-expose) | tidak di-expose ke host | keamanan |
| Image tag | `:latest` saja | `:<git-sha>` + `:latest` | bisa rollback |
| Storage | — | R2 (env `R2_*`) | upload foto |

---

## 1. File di repo (buat dulu — undangin belum punya)

### `Dockerfile` (undangin-api)
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/migrate .
EXPOSE 8088
CMD ["./api"]
```
> Migrasi auto-apply saat boot (`internal/app/app.go`). Binary `migrate` disertakan
> hanya untuk operasi manual (mis. `migrate down` saat troubleshooting).

### `docker-compose.yml` (undangin-api)
```yaml
services:
  postgres:
    image: postgres:18-alpine
    container_name: undangin_db
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - undangin_pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: always
    # NOTE: sengaja TIDAK expose 5432 ke host. API akses via network compose.
    # Kalau perlu psql dari luar, pakai SSH tunnel, atau bind localhost saja:
    # ports: ["127.0.0.1:5432:5432"]

  undangin-api:
    image: ghcr.io/irvanmhndra/siaga-api:${IMAGE_TAG:-latest}
    container_name: undangin_api
    env_file: .env
    ports:
      - "127.0.0.1:8088:8088"   # cuma nginx yang perlu; jangan expose publik
    volumes:
      - ./web-dist:/app/web:ro  # dist dari undangin-web (di-scp CI)
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost:8088/health || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always

volumes:
  undangin_pgdata:
```

---

## 2. VPS initial setup (one-time)

Sebagai user `deploy`:
```bash
mkdir -p ~/undangin/web-dist
```

Buat `~/undangin/.env` (chmod 600):
```bash
tee ~/undangin/.env > /dev/null <<'EOF'
# Server
SERVER_PORT=8088
SERVER_ENV=production

# PostgreSQL
POSTGRES_HOST=undangin_db
POSTGRES_PORT=5432
POSTGRES_USER=undangin_user
POSTGRES_PASSWORD=__isi_password_kuat__
POSTGRES_DB=undangin_db
POSTGRES_SSLMODE=disable

# Serve SPA + OG injection dari container ini
WEB_DIST=/app/web

# Cloudflare R2 (upload foto) — dari setup R2 yang sudah kelar
R2_ACCOUNT_ID=__account_id__
R2_ACCESS_KEY_ID=__access_key__
R2_SECRET_ACCESS_KEY=__secret__
R2_BUCKET=undangin-prod
R2_ENDPOINT=https://__account_id__.r2.cloudflarestorage.com
R2_PUBLIC_BASE_URL=https://undangin-cdn.irvanmahendra.com
EOF
chmod 600 ~/undangin/.env
```
> Tidak ada `JWT_SECRET` — Undangin pakai opaque token (tabel `user_sessions`).

Copy `docker-compose.yml` ke VPS — **jalankan dari laptop, di dalam folder repo `undangin-api`** (bukan dari VPS):
```bash
scp docker-compose.yml deploy@<VPS_IP>:~/undangin/
```
Login GHCR di VPS (interaktif — jangan echo token ke shell/history):
```bash
docker login ghcr.io -u irvanmhndra
# tempel Personal Access Token (scope read:packages) saat diminta "Password"
```
> ⚠️ **JANGAN `docker compose up` manual di first deploy** — image belum ada di GHCR.
> Image pertama dibuat oleh CI saat push ke `main`, dan CI juga yang jalanin `up` di VPS.
> Lihat **Urutan first deploy** di bawah.

---

## 3. Nginx (single-origin, as root)

nginx = pintu depan (port 80/443). App Go cuma dengerin di `127.0.0.1:8088` (internal).
Blok `server` di bawah nyuruh nginx nerusin request `undangin.irvanmahendra.com` ke Go
(reverse proxy). Snippet ini **DISIMPAN JADI FILE**, bukan dijalanin langsung.
SSH ke VPS sebagai root, lalu copy-paste satu blok ini:

```bash
# 1. Tulis config ke file
sudo tee /etc/nginx/sites-available/undangin > /dev/null <<'EOF'
server {
    listen 80;
    server_name undangin.irvanmahendra.com;

    client_max_body_size 5M;   # presigned upload langsung ke R2, jadi cukup kecil

    location / {
        proxy_pass http://127.0.0.1:8088;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

# 2. Aktifkan (symlink ke sites-enabled) + tes + reload
sudo ln -s /etc/nginx/sites-available/undangin /etc/nginx/sites-enabled/undangin
sudo nginx -t && sudo systemctl reload nginx
```

## 4. HTTPS (certbot, as root)
```bash
certbot --nginx -d undangin.irvanmahendra.com
```

## 5. DNS (Cloudflare — DNS only / grey cloud, ikut pola app lain)
```
undangin.irvanmahendra.com   A   <VPS_IP>   (DNS only)
```
> `cdn.undangin...` sudah ada (custom domain R2, proxied — dikelola R2).

---

## 6. GitHub Actions

Secrets di **kedua** repo (Settings → Secrets → Actions): `VPS_HOST`, `VPS_USER`
(`deploy`), `VPS_SSH_KEY`. (`GITHUB_TOKEN` untuk GHCR sudah otomatis.)

### `undangin-api/.github/workflows/deploy.yml`
```yaml
name: Deploy undangin-api
on:
  push:
    branches: [main]
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - uses: actions/setup-go@v5
        with: { go-version: '1.25.x' }
      - uses: golangci/golangci-lint-action@v7
        with: { version: v2.11.3, args: --timeout=5m }
      - run: go test -race -timeout 120s ./...
      # (opsional, seperti nexpos) tambah govulncheck + gosec kalau mau

  deploy:
    needs: [ci]
    runs-on: ubuntu-latest
    permissions: { contents: read, packages: write }
    steps:
      - uses: actions/checkout@v5
      - name: Login GHCR
        run: echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
      - name: Build & push (SHA + latest)
        run: |
          IMG=ghcr.io/irvanmhndra/siaga-api
          docker build -t $IMG:${{ github.sha }} -t $IMG:latest .
          docker push $IMG:${{ github.sha }}
          docker push $IMG:latest
      - name: Deploy (pin SHA)
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd ~/undangin
            export IMAGE_TAG=${{ github.sha }}
            docker compose pull undangin-api
            docker compose up -d undangin-api
            docker image prune -f          # dangling saja — image ber-SHA disimpan utk rollback
            sleep 5
            wget -qO- http://127.0.0.1:8088/health && echo " ✅ deployed ${{ github.sha }}"
```
> Migrasi auto-apply saat container boot — tak ada step migrate terpisah.
> Coverage gate 50% ala nexpos SENGAJA tidak dipasang: test undangin masih tipis;
> tambahkan gate saat coverage sudah naik.

### `undangin-web/.github/workflows/deploy.yml`
```yaml
name: Deploy undangin-web
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20, cache: npm }
      - run: npm ci
      - run: npm run build          # baca .env.production
      - name: Ship dist ke VPS
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          source: dist/
          target: ~/undangin/web-dist/
          strip_components: 1
      - name: Restart API (reload index.html yang di-cache)
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: cd ~/undangin && docker compose restart undangin-api
```
> `SPAHandler` meng-cache `index.html` di memori saat start, jadi web deploy perlu
> restart container biar `index.html` baru kebaca. (Alternatif: ubah handler baca
> `index.html` per-request → tak perlu restart. Kecil, bisa nyusul.)

### `undangin-web/.env.production` (di-commit — bukan rahasia)
```
VITE_API_URL=https://undangin.irvanmahendra.com/api/v1
```
> Single-origin: web panggil API di origin yang sama. (`config/api.ts` baca `VITE_API_URL`.)

---

## 7. Backup database (penting — ada data tamu/undangan)

> "AWS CLI" / "rclone" di sini **bukan** berarti pakai AWS. R2 itu S3-compatible,
> jadi tool S3 apa pun bisa ngomong ke R2. VPS-mu tetap Contabo — nggak ada AWS.

### A. Backup lokal (mulai dari sini — nol dependency)
Dump DB ke file di VPS + buang yang lebih tua dari 7 hari:
```bash
# ~/undangin/backup.sh
#!/bin/bash
mkdir -p ~/undangin/backups
docker exec undangin_db pg_dump -U undangin_user undangin_db \
  | gzip > ~/undangin/backups/undangin-$(date +%F).sql.gz
find ~/undangin/backups -name '*.sql.gz' -mtime +7 -delete
```
```bash
chmod +x ~/undangin/backup.sh
crontab -e     # tambah baris ini (tiap hari jam 3 pagi):
0 3 * * * /home/deploy/undangin/backup.sh
```
Restore: `gunzip -c undangin-<tgl>.sql.gz | docker exec -i undangin_db psql -U undangin_user -d undangin_db`

### B. (Opsional, nanti) Salin offsite ke R2 pakai rclone
Kalau VPS mati, backup lokal ikut hilang — makanya disalin ke R2. Pakai **rclone**
(lebih simpel dari aws CLI): sekali `rclone config` (provider: Cloudflare R2), lalu
tambah di akhir `backup.sh`:
```bash
rclone copy ~/undangin/backups r2:undangin-backups --max-age 25h
```
⚠️ Backup **jangan** ke bucket `undangin-prod` (itu publik lewat `cdn.`). Bikin bucket
**privat terpisah** `undangin-backups` (tanpa custom domain).

---

## Urutan first deploy (chicken-and-egg: image harus ada dulu)

1. **DNS** — `undangin.irvanmahendra.com A <VPS_IP>` (DNS only) di Cloudflare.
2. **VPS prep** — `mkdir -p ~/undangin/web-dist`; buat `~/undangin/.env` (bagian 2);
   scp `docker-compose.yml` **dari laptop**; `docker login ghcr.io` (token baru, read:packages).
   **Jangan `docker compose up` dulu.**
3. **nginx + certbot** (bagian 3 & 4). Sementara bakal 502 karena app belum jalan — normal.
4. **GitHub secrets** (`VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`) di kedua repo.
5. **Push `undangin-api` ke `main`** → CI: lint+test → build image → push GHCR → ssh ke VPS
   `docker compose up` (image pertama lahir + container start). 502 hilang.
6. **Push `undangin-web` ke `main`** → CI: build → scp `dist` → restart API. SPA live.

Setelah ini, tiap push ke `main` auto-deploy.

## Rollback
```bash
cd ~/undangin
export IMAGE_TAG=<git-sha-lama>   # SHA versi bagus sebelumnya
docker compose pull undangin-api && docker compose up -d undangin-api
```

## Useful commands
```bash
docker logs undangin_api -f
docker compose -f ~/undangin/docker-compose.yml restart undangin-api
docker exec -it undangin_db psql -U undangin_user -d undangin_db
```
