-- +goose Up
-- +goose StatementBegin
create table public.t_pass
(
    id         uuid default gen_random_uuid()
        constraint t_pass_pk
            primary key,
    user_id    bigint
        constraint t_pass_t_user_id_fk
            references public.t_user,
    domen      varchar,
    username   bytea,
    password   bytea,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
comment on table public.t_pass is 'Таблица с хранением логинов/паролей';
comment on column public.t_pass.user_id is 'Владелец связки';
comment on column public.t_pass.domen is 'Домен, к которому относися связка';
comment on column public.t_pass.username is 'Логин';
comment on column public.t_pass.password is 'Пароль';
comment on column public.t_pass.created_at is 'Дата создания';
comment on column public.t_pass.updated_at is 'Дата обновления';
create index t_pass_user_id_index
    on public.t_pass (user_id);


create table public.t_comment
(
    id           uuid default gen_random_uuid()
        constraint t_comment_pk
            primary key,
    content_type smallint,
    content_id   uuid,
    comment      varchar,
    created_at   timestamp default now(),
    updated_at   timestamp default now()
);
comment on table public.t_comment is 'Комментарии к хранимому контенту';
comment on column public.t_comment.content_type is 'К какому типу относится комментарий';
comment on column public.t_comment.content_id is 'Идентификатор к которому относится комментарий';
comment on column public.t_comment.comment is 'Комментарий с метаинформацией';
create index t_comment_content_id_content_type_index
    on public.t_comment (content_id, content_type);



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
