-- +goose Up
-- +goose StatementBegin
create table public.t_text
(
    id         uuid default gen_random_uuid()
        constraint t_text_pk
            primary key,
    user_id    bigint
        constraint t_text_t_user_id_fk
            references public.t_user,
    text_data  bytea,
    created_at timestamp default now(),
    updated_at timestamp default now()
);

comment on table public.t_text is 'Произвольные зашифрованные текстовые даныне';

comment on column public.t_text.user_id is 'Идентификатор владельца ползователя';

comment on column public.t_text.text_data is 'Зашифрованные текстовые данные';

create index t_text_user_id_index
    on public.t_text (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
