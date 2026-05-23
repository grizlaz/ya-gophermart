-- +goose Up
create table if not exists "user" (
	id serial primary key,
	login text not null,
	password bytea not null 
);

create index if not exists idx_user_login on "user"(login);



create table if not exists "order" (
	"number" text primary key,
	user_id int references "user"(id),
	status text not null,
	accrual int,
	created_at timestamptz default now(),
	updated_at timestamptz default now()
);

create index if not exists idx_order_user_id on "order"(user_id);


create table if not exists "wallet" (
	user_id int primary key references "user"(id),
	balance int not null,
	withdrawn int not null
);


create table if not exists "wallet_history" (
	user_id int references "wallet"(user_id),
	"number" text not null,
	operation int not null,
	amount int not null,
	date timestamptz default now(),
	PRIMARY KEY(user_id, date)
);

-- +goose Down
drop table if exists "user";
drop table if exists "order";
drop table if exists "wallet";
drop table if exists "wallet_history";