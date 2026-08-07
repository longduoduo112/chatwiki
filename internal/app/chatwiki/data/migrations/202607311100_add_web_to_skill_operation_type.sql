-- +goose Up

ALTER TABLE "public"."chat_ai_web_to_skill_task"
    ADD COLUMN "operation_type" int2 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_web_to_skill_task"."operation_type" IS '任务操作类型:0生成,1更新';

-- +goose Down

ALTER TABLE "public"."chat_ai_web_to_skill_task"
    DROP COLUMN IF EXISTS "operation_type";
