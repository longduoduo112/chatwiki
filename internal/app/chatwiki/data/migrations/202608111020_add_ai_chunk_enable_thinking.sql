-- +goose Up

ALTER TABLE "public"."chat_ai_library"
    ADD COLUMN "ai_chunk_enable_thinking" int2 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_library"."ai_chunk_enable_thinking" IS 'AI分段深度思考开关:0关,1开';

ALTER TABLE "public"."chat_ai_library_file"
    ADD COLUMN "ai_chunk_enable_thinking" int2 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_library_file"."ai_chunk_enable_thinking" IS 'AI分段深度思考开关:0关,1开';

-- +goose Down

ALTER TABLE "public"."chat_ai_library_file"
    DROP COLUMN "ai_chunk_enable_thinking";

ALTER TABLE "public"."chat_ai_library"
    DROP COLUMN "ai_chunk_enable_thinking";
