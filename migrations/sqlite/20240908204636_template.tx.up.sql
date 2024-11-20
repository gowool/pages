CREATE TABLE "templates" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "name" TEXT NOT NULL,
    "content" TEXT NOT NULL DEFAULT '',
    "enabled" BOOLEAN NOT NULL DEFAULT FALSE,
    "created" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "updated" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f'))
);

--bun:split

CREATE INDEX "templates_created_updated_idx" ON "templates" ("created", "updated");
CREATE INDEX "templates_name_enabled_idx" ON "templates" ("name", "enabled");
