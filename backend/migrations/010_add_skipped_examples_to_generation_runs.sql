begin;

alter table learning.generation_runs
add column if not exists skipped_examples integer not null default 0;

commit;
