SET statement_timeout = 0;

--==============================================================================
--bun:split

CREATE TABLE "templates" (
    "id" integer PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    "name" varchar NOT NULL,
    "content" varchar NOT NULL DEFAULT '',
    "enabled" boolean NOT NULL DEFAULT false,
    "created" timestamptz NOT NULL DEFAULT now(),
    "updated" timestamptz NOT NULL DEFAULT now()
);

--bun:split

CREATE INDEX "templates_created_updated_idx" ON "templates" ("created", "updated");
CREATE INDEX "templates_name_enabled_idx" ON "templates" ("name", "enabled");