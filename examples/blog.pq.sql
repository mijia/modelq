begin;

DROP TABLE "article";
DROP TABLE "user";

CREATE TABLE "user" (
    "id" BIGSERIAL,
    "name" VARCHAR(50) NOT NULL,
    "password" VARCHAR(50) NOT NULL,
    "is_married" BOOLEAN DEFAULT NULL,
    "age" INT DEFAULT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("name")
);

CREATE INDEX "INDEX_user_age" ON "user" ("age" ASC);

CREATE TABLE "article" (
    "id" BIGSERIAL,
    "user_id" BIGINT NOT NULL,
    "title" VARCHAR(512) NOT NULL,
    "state" SMALLINT NOT NULL DEFAULT 0,  -- "0: published, 1: draft, 2: hidden",
    "content" TEXT DEFAULT NULL,
    "donation" DECIMAL(12, 2) DEFAULT 0.5,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE CASCADE
);

CREATE INDEX "INDEX_article_state" ON "article" ("state");

commit;