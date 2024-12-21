-- +goose Up
-- +goose StatementBegin
create table public.t_file
(
    id         bigserial
        constraint t_file_pk
            primary key,
    name       bytea,
    created_at  timestamp default now(),
    updated_at timestamp default now(),
    user_id    bigint
        constraint t_file_t_user_id_fk
            references public.t_user,
    file_path  varchar
);
comment on table public.t_file is 'Данные о файлах';
comment on column public.t_file.name is 'Название оригинального файла';
comment on column public.t_file.user_id is 'Владелец файла';
comment on column public.t_file.file_path is 'путь к файлу';
create index t_file_user_id_index
    on public.t_file (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
