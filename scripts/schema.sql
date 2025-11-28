-- Basic schema
CREATE TABLE IF NOT EXISTS facilities(
  id serial primary key,
  name text not null
);
CREATE TABLE IF NOT EXISTS meters(
  id serial primary key,
  facility_id int not null references facilities(id),
  serial text not null
);
CREATE TABLE IF NOT EXISTS readings(
  id bigserial primary key,
  meter_id int not null references meters(id),
  timestamp timestamptz not null,
  voltage double precision not null,
  current double precision not null,
  power_kw double precision not null
);
