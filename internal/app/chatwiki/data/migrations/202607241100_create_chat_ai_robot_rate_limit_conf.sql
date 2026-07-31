-- +goose Up

CREATE TABLE "public"."chat_ai_robot_rate_limit_conf"
(
    "id"                        serial      NOT NULL primary key,
    "admin_user_id"             int4        NOT NULL DEFAULT 0,
    "robot_key"                 varchar(15) NOT NULL DEFAULT '',
    "switch_status"             int2        NOT NULL DEFAULT 0,
    "five_minute_limit"         int8        NOT NULL DEFAULT 10,
    "five_minute_reply_type"    int2        NOT NULL DEFAULT 1,
    "five_minute_reply_content" text        NOT NULL DEFAULT '',
    "daily_limit"               int8        NOT NULL DEFAULT 100,
    "daily_reply_type"          int2        NOT NULL DEFAULT 0,
    "daily_reply_content"       text        NOT NULL DEFAULT '',
    "create_time"               int4        NOT NULL DEFAULT 0,
    "update_time"               int4        NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX ON "public"."chat_ai_robot_rate_limit_conf" ("robot_key");
CREATE INDEX ON "public"."chat_ai_robot_rate_limit_conf" ("admin_user_id");

COMMENT ON TABLE "public"."chat_ai_robot_rate_limit_conf" IS '机器人消息限流配置表';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."id" IS 'ID';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."admin_user_id" IS '管理员用户ID';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."robot_key" IS '机器人key';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."switch_status" IS '限流开关:0关,1开';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."five_minute_limit" IS '同一用户5分钟消息上限';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."five_minute_reply_type" IS '5分钟限流处理:0不回复,1回复指定文案';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."five_minute_reply_content" IS '5分钟限流回复文案';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."daily_limit" IS '同一用户自然日消息上限';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."daily_reply_type" IS '天级限流处理:0不回复,1回复指定文案';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."daily_reply_content" IS '天级限流回复文案';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."create_time" IS '创建时间';
COMMENT ON COLUMN "public"."chat_ai_robot_rate_limit_conf"."update_time" IS '更新时间';

-- +goose Down

DROP TABLE IF EXISTS "public"."chat_ai_robot_rate_limit_conf";
