-- +goose Up
-- +goose StatementBegin
create table public.t_card
(
    id      uuid default gen_random_uuid()
        constraint t_card_pk
            primary key,
    number  bytea,
    date    bytea,
    owner   bytea,
    cvv     bytea,
    user_id bigint
        constraint t_card_t_user_id_fk
            references public.t_user,
    created_at timestamp default now(),
    updated_at timestamp default now()
);
comment on table public.t_card is 'Хранение данных банковских карт';
comment on column public.t_card.number is 'Номер карты';
comment on column public.t_card.date is 'Дата дейтсвия';
comment on column public.t_card.owner is 'Имя на карте';
comment on column public.t_card.cvv is 'Код карты';
comment on column public.t_card.user_id is 'Владелец записи';
create index t_card_user_id_index
    on public.t_card (user_id);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd
