-- +goose Up

ALTER TABLE "public"."chat_ai_library"
    ADD COLUMN "is_permanent" int2 NOT NULL DEFAULT 1,
    ADD COLUMN "expire_time" int8 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_library"."is_permanent" IS '是否永久有效 0:否 1:是';
COMMENT ON COLUMN "public"."chat_ai_library"."expire_time" IS '有效期截止时间，秒级Unix时间戳，永久有效时为0';

-- +goose Down

ALTER TABLE "public"."chat_ai_library"
    DROP COLUMN "expire_time",
    DROP COLUMN "is_permanent";
