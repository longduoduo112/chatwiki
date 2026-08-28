-- +goose Up

ALTER TABLE "public"."chat_ai_goods_library"
    ADD COLUMN "goods_wechat_card" jsonb NOT NULL DEFAULT '{}';

COMMENT ON COLUMN "public"."chat_ai_goods_library"."goods_wechat_card" IS '商品微信小程序卡片配置';

-- +goose Down

ALTER TABLE "public"."chat_ai_goods_library"
    DROP COLUMN IF EXISTS "goods_wechat_card";
