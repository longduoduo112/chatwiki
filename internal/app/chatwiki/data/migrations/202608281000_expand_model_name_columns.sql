-- +goose Up

ALTER TABLE "public"."llm_token_daily_stats"
    ALTER COLUMN "model" TYPE varchar(100);

ALTER TABLE "public"."llm_token_app_daily_stats"
    ALTER COLUMN "model" TYPE varchar(100);

ALTER TABLE "public"."chat_ai_robot"
    ALTER COLUMN "optimize_question_use_model" TYPE varchar(100);

ALTER TABLE "public"."chat_ai_model_list"
    ALTER COLUMN "show_model_name" TYPE varchar(100);
