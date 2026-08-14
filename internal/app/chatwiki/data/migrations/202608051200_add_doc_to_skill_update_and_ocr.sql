-- +goose Up

ALTER TABLE "public"."chat_ai_doc_to_skill_task"
    ADD COLUMN "pending_source_files" text NOT NULL DEFAULT '',
    ADD COLUMN "operation_type" int2 NOT NULL DEFAULT 0,
    ADD COLUMN "online_ocr" int2 NOT NULL DEFAULT 0,
    ADD COLUMN "version" int4 NOT NULL DEFAULT 1;

COMMENT ON COLUMN "public"."chat_ai_doc_to_skill_task"."pending_source_files" IS '更新任务本次新增的源文档列表(JSON)';
COMMENT ON COLUMN "public"."chat_ai_doc_to_skill_task"."operation_type" IS '任务操作类型:0生成,1更新';
COMMENT ON COLUMN "public"."chat_ai_doc_to_skill_task"."online_ocr" IS '是否对PDF启用在线OCR:0否,1是';
COMMENT ON COLUMN "public"."chat_ai_doc_to_skill_task"."version" IS '当前最新成功技能包版本号';

-- +goose Down

ALTER TABLE "public"."chat_ai_doc_to_skill_task"
    DROP COLUMN IF EXISTS "version",
    DROP COLUMN IF EXISTS "online_ocr",
    DROP COLUMN IF EXISTS "operation_type",
    DROP COLUMN IF EXISTS "pending_source_files";
