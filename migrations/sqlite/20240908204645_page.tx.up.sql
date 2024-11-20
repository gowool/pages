CREATE TABLE "pages" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "parent_id" INTEGER,
    "site_id" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "title" TEXT,
    "pattern" TEXT NOT NULL DEFAULT '_page_cms',
    "alias" TEXT,
    "slug" TEXT,
    "url" TEXT,
    "custom_url" TEXT,
    "template" TEXT NOT NULL,
    "position" INTEGER NOT NULL DEFAULT 0,
    "decorate" BOOLEAN NOT NULL DEFAULT TRUE,
    "javascript" TEXT,
    "stylesheet" TEXT,
    "headers" JSON NOT NULL DEFAULT '{}',
    "metas" JSON NOT NULL DEFAULT '[]',
    "metadata" JSON NOT NULL DEFAULT '{}',
    "created" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "updated" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "published" TEXT,
    "expired" TEXT,
    ---
    FOREIGN KEY ("parent_id") REFERENCES "pages" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
    FOREIGN KEY ("site_id") REFERENCES "sites" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);

--bun:split

CREATE INDEX "pages_created_updated_idx" ON "pages" ("created", "updated");
CREATE INDEX "pages_published_expired_idx" ON "pages" ("published", "expired");
CREATE INDEX "pages_pattern_idx" ON "pages" ("pattern");
CREATE INDEX "pages_alias_idx" ON "pages" ("alias");
CREATE INDEX "pages_slug_idx" ON "pages" ("slug");
CREATE INDEX "pages_url_idx" ON "pages" ("url");
CREATE INDEX "pages_custom_url_idx" ON "pages" ("custom_url");
CREATE INDEX "pages_position_idx" ON "pages" ("position");
