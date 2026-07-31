-- +goose Up

ALTER TABLE "public"."form_field_value"
    ALTER COLUMN "integer_content" TYPE int8,
    ALTER COLUMN "number_content" TYPE float8;

-- +goose Down

ALTER TABLE "public"."form_field_value"
    ALTER COLUMN "integer_content" TYPE int4,
    ALTER COLUMN "number_content" TYPE float4;
