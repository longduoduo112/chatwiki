-- +goose Up

ALTER TABLE "public"."chat_ai_faq_files"
    ADD COLUMN "chunk_enable_thinking" int2 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_faq_files"."chunk_enable_thinking" IS 'FAQ提取深度思考开关:0关,1开';

-- +goose Down

ALTER TABLE "public"."chat_ai_faq_files"
    DROP COLUMN "chunk_enable_thinking";
