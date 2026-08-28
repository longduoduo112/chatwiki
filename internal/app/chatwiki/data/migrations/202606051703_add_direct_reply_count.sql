-- +goose Up
-- AI推荐回复-原文模式/传统模式支持设置回复数量
ALTER TABLE "public"."chat_ai_robot"
ADD COLUMN "library_qa_direct_reply_count" int2 DEFAULT 1;
COMMENT ON COLUMN "public"."chat_ai_robot"."library_qa_direct_reply_count" IS '知识库QA直连回复数量(仅知识库模式)';

ALTER TABLE "public"."chat_ai_robot"
ADD COLUMN "mixture_qa_direct_reply_count" int2 DEFAULT 1;
COMMENT ON COLUMN "public"."chat_ai_robot"."mixture_qa_direct_reply_count" IS '混合模式QA直连回复数量';

ALTER TABLE "public"."chat_ai_message"
ADD COLUMN "answer_list" text DEFAULT '';
COMMENT ON COLUMN "public"."chat_ai_message"."answer_list" IS '多答案列表JSON（QA直连回复模式，每条含图/视频markdown）';
