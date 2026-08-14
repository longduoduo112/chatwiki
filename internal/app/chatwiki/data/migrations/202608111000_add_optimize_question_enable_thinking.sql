-- +goose Up

ALTER TABLE "public"."chat_ai_robot"
    ADD COLUMN "optimize_question_enable_thinking" int2 NOT NULL DEFAULT 0;

COMMENT ON COLUMN "public"."chat_ai_robot"."optimize_question_enable_thinking" IS '问题优化深度思考开关:0关,1开';

-- +goose Down

ALTER TABLE "public"."chat_ai_robot"
    DROP COLUMN "optimize_question_enable_thinking";
