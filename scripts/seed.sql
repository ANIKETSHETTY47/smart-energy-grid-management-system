insert into facilities(name) values ('Plant A') on conflict do nothing;
insert into meters(facility_id, serial) values (1,'meter-001') on conflict do nothing;
