CREATE TABLE "sequence_nodes" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL
);

--==============================================================================
--bun:split

CREATE TABLE "nodes" (
    "id" INTEGER PRIMARY KEY NOT NULL,
    "parent_id" INTEGER NOT NULL DEFAULT 0,
    "name" TEXT NOT NULL,
    "label" TEXT,
    "uri" TEXT,
    "path" TEXT NOT NULL,
    "level" INTEGER NOT NULL DEFAULT 0,
    "position" INTEGER NOT NULL DEFAULT 0,
    "display_children" BOOLEAN NOT NULL DEFAULT TRUE,
    "display" BOOLEAN NOT NULL DEFAULT TRUE,
    "attributes" JSON NOT NULL DEFAULT '{}',
    "link_attributes" JSON NOT NULL DEFAULT '{}',
    "children_attributes" JSON NOT NULL DEFAULT '{}',
    "label_attributes" JSON NOT NULL DEFAULT '{}',
    "metadata" JSON NOT NULL DEFAULT '{}',
    "created" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    "updated" TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f')),
    ---
    FOREIGN KEY ("id") REFERENCES "sequence_nodes" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);

--bun:split

CREATE INDEX "nodes_created_updated_idx" ON "nodes" ("created", "updated");
CREATE INDEX "nodes_path_idx" ON "nodes" ("path");
CREATE INDEX "nodes_level_idx" ON "nodes" ("level");
