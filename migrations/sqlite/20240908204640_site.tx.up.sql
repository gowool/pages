CREATE TABLE "sites" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name" TEXT NOT NULL,
    "title" TEXT,
    "separator" TEXT NOT NULL DEFAULT ' - ',
    "host" TEXT NOT NULL,
    "locale" TEXT,
    "relative_path" TEXT,
    "is_default" BOOLEAN NOT NULL DEFAULT FALSE,
    "javascript" TEXT,
    "stylesheet" TEXT,
    "metas" JSON NOT NULL DEFAULT '[]',
    "metadata" JSON NOT NULL DEFAULT '{}',
    "created" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "updated" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "published" TEXT,
    "expired" TEXT
);

--bun:split

CREATE INDEX "sites_created_updated_idx" ON "sites" ("created", "updated");
CREATE INDEX "sites_published_expired_idx" ON "sites" ("published", "expired");
CREATE INDEX "sites_host_is_default_idx" ON "sites" ("host", "is_default");

--bun:split

CREATE UNIQUE INDEX "sites_host_lifespan_unq" ON "sites" (lower("host"),
                                                          lower(coalesce("locale", '')),
                                                          lower(coalesce("relative_path", '')),
                                                          coalesce("published", '1970-01-01 00:00:00.000'),
                                                          coalesce("expired", '1970-01-01 00:00:00.000'));
