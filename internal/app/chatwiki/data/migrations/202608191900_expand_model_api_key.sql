-- +goose Up

ALTER TABLE "public"."chat_ai_model_config"
    ALTER COLUMN "api_key" TYPE text;
