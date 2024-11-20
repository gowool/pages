CREATE TABLE "menus" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    "node_id" INTEGER,
    "name" TEXT NOT NULL,
    "handle" TEXT NOT NULL ,
    "enabled" BOOLEAN NOT NULL DEFAULT FALSE,
    "created" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "updated" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    ---
    FOREIGN KEY ("node_id") REFERENCES "nodes" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);

--bun:split

CREATE INDEX "menus_created_updated_idx" ON "menus" ("created", "updated");
CREATE INDEX "menus_enabled_idx" ON "menus" ("enabled");

--bun:split

CREATE UNIQUE INDEX "menus_handle_unq" ON "menus" (lower("handle"));
